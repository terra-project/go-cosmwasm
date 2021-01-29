mod api;
mod db;
mod error;
mod gas_meter;
mod iterator;
mod memory;
mod querier;
mod tests;

pub use api::GoApi;
pub use db::{db_t, DB};
pub use memory::{V3_free_rust, Buffer};
pub use querier::GoQuerier;

use std::convert::TryInto;
use std::panic::{catch_unwind, AssertUnwindSafe};
use std::str::from_utf8;

use crate::error::{clear_error, handle_c_error, set_error, Error};
use cosmwasm_vm::{
    call_handle_raw, call_init_raw, call_migrate_raw, call_query_raw, features_from_csv, Checksum,
    Cache, Extern, CacheOptions, Size
};

const MEMORY_CACHE_SIZE: Size = Size::mebi(500); // TODO: Make configurable

#[repr(C)]
pub struct cache_t {}

fn to_cache(ptr: *mut cache_t) -> Option<&'static mut Cache<DB, GoApi, GoQuerier>> {
    if ptr.is_null() {
        None
    } else {
        let c = unsafe { &mut *(ptr as *mut Cache<DB, GoApi, GoQuerier>) };
        Some(c)
    }
}

fn to_extern(storage: DB, api: GoApi, querier: GoQuerier) -> Extern<DB, GoApi, GoQuerier> {
    Extern {
        storage,
        api,
        querier,
    }
}

#[no_mangle]
pub extern "C" fn V3_init_cache(
    data_dir: Buffer,
    supported_features: Buffer,
    // TODO: remove unused cache size
    _cache_size: usize,
    err: Option<&mut Buffer>,
) -> *mut cache_t {
    let r = catch_unwind(|| do_init_cache(data_dir, supported_features))
        .unwrap_or_else(|_| Err(Error::panic()));
    match r {
        Ok(t) => {
            clear_error();
            t as *mut cache_t
        }
        Err(e) => {
            set_error(e, err);
            std::ptr::null_mut()
        }
    }
}

// store some common string for argument names
static DATA_DIR_ARG: &str = "data_dir";
static FEATURES_ARG: &str = "supported_features";
static CACHE_ARG: &str = "cache";
static WASM_ARG: &str = "wasm";
static CODE_ID_ARG: &str = "code_id";
static MSG_ARG: &str = "msg";
static PARAMS_ARG: &str = "params";
static GAS_USED_ARG: &str = "gas_used";

fn do_init_cache(
    data_dir: Buffer,
    supported_features: Buffer,
) -> Result<*mut Cache<DB, GoApi, GoQuerier>, Error> {
    let dir = unsafe { data_dir.read() }.ok_or_else(|| Error::empty_arg(DATA_DIR_ARG))?;
    let dir_str = from_utf8(dir)?;
    // parse the supported features
    let features_bin =
        unsafe { supported_features.read() }.ok_or_else(|| Error::empty_arg(FEATURES_ARG))?;
    let features_str = from_utf8(features_bin)?;
    let features = features_from_csv(features_str);
    let options = CacheOptions {
        base_dir: dir_str.into(),
        supported_features: features,
        memory_cache_size: MEMORY_CACHE_SIZE,
    };
    let cache = unsafe { Cache::new(options) }?;
    let out = Box::new(cache);
    Ok(Box::into_raw(out))
}

/// frees a cache reference
///
/// # Safety
///
/// This must be called exactly once for any `*cache_t` returned by `init_cache`
/// and cannot be called on any other pointer.
#[no_mangle]
pub extern "C" fn V3_release_cache(cache: *mut cache_t) {
    if !cache.is_null() {
        // this will free cache when it goes out of scope
        let _ = unsafe { Box::from_raw(cache as *mut Cache<DB, GoApi, GoQuerier>) };
    }
}

#[no_mangle]
pub extern "C" fn V3_create(cache: *mut cache_t, wasm: Buffer, err: Option<&mut Buffer>) -> Buffer {
    let r = match to_cache(cache) {
        Some(c) => catch_unwind(AssertUnwindSafe(move || do_create(c, wasm)))
            .unwrap_or_else(|_| Err(Error::panic())),
        None => Err(Error::empty_arg(CACHE_ARG)),
    };
    let data = handle_c_error(r, err);
    Buffer::from_vec(data)
}

fn do_create(cache: &mut Cache<DB, GoApi, GoQuerier>, wasm: Buffer) -> Result<Checksum, Error> {
    let wasm = unsafe { wasm.read() }.ok_or_else(|| Error::empty_arg(WASM_ARG))?;
    let checksum = cache.save_wasm(wasm)?;
    Ok(checksum)
}

#[no_mangle]
pub extern "C" fn V3_get_code(cache: *mut cache_t, id: Buffer, err: Option<&mut Buffer>) -> Buffer {
    let r = match to_cache(cache) {
        Some(c) => catch_unwind(AssertUnwindSafe(move || do_get_code(c, id)))
            .unwrap_or_else(|_| Err(Error::panic())),
        None => Err(Error::empty_arg(CACHE_ARG)),
    };
    let data = handle_c_error(r, err);
    Buffer::from_vec(data)
}

fn do_get_code(cache: &mut Cache<DB, GoApi, GoQuerier>, id: Buffer) -> Result<Vec<u8>, Error> {
    let id: Checksum = unsafe { id.read() }
        .ok_or_else(|| Error::empty_arg(CACHE_ARG))?
        .try_into()?;
    let wasm = cache.load_wasm(&id)?;
    Ok(wasm)
}

#[no_mangle]
pub extern "C" fn V3_instantiate(
    cache: *mut cache_t,
    contract_id: Buffer,
    params: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
    err: Option<&mut Buffer>,
) -> Buffer {
    let r = match to_cache(cache) {
        Some(c) => catch_unwind(AssertUnwindSafe(move || {
            do_init(
                c,
                contract_id,
                params,
                msg,
                db,
                api,
                querier,
                gas_limit,
                gas_used,
            )
        }))
        .unwrap_or_else(|_| Err(Error::panic())),
        None => Err(Error::empty_arg(CACHE_ARG)),
    };
    let data = handle_c_error(r, err);
    Buffer::from_vec(data)
}

fn do_init(
    cache: &mut Cache<DB, GoApi, GoQuerier>,
    code_id: Buffer,
    params: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
) -> Result<Vec<u8>, Error> {
    let gas_used = gas_used.ok_or_else(|| Error::empty_arg(GAS_USED_ARG))?;
    let code_id: Checksum = unsafe { code_id.read() }
        .ok_or_else(|| Error::empty_arg(CODE_ID_ARG))?
        .try_into()?;
    let params = unsafe { params.read() }.ok_or_else(|| Error::empty_arg(PARAMS_ARG))?;
    let msg = unsafe { msg.read() }.ok_or_else(|| Error::empty_arg(MSG_ARG))?;

    let deps = to_extern(db, api, querier);
    let mut instance = cache.get_instance(&code_id, deps, gas_limit)?;
    // We only check this result after reporting gas usage and returning the instance into the cache.
    let res = call_init_raw(&mut instance, params, msg);
    *gas_used = instance.create_gas_report().used_internally;
    instance.recycle();
    Ok(res?)
}

#[no_mangle]
pub extern "C" fn V3_handle(
    cache: *mut cache_t,
    code_id: Buffer,
    params: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
    err: Option<&mut Buffer>,
) -> Buffer {
    let r = match to_cache(cache) {
        Some(c) => catch_unwind(AssertUnwindSafe(move || {
            do_handle(
                c, code_id, params, msg, db, api, querier, gas_limit, gas_used,
            )
        }))
        .unwrap_or_else(|_| Err(Error::panic())),
        None => Err(Error::empty_arg(CACHE_ARG)),
    };
    let data = handle_c_error(r, err);
    Buffer::from_vec(data)
}

fn do_handle(
    cache: &mut Cache<DB, GoApi, GoQuerier>,
    code_id: Buffer,
    params: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
) -> Result<Vec<u8>, Error> {
    let gas_used = gas_used.ok_or_else(|| Error::empty_arg(GAS_USED_ARG))?;
    let code_id: Checksum = unsafe { code_id.read() }
        .ok_or_else(|| Error::empty_arg(CODE_ID_ARG))?
        .try_into()?;
    let params = unsafe { params.read() }.ok_or_else(|| Error::empty_arg(PARAMS_ARG))?;
    let msg = unsafe { msg.read() }.ok_or_else(|| Error::empty_arg(MSG_ARG))?;

    let deps = to_extern(db, api, querier);
    let mut instance = cache.get_instance(&code_id, deps, gas_limit)?;
    // We only check this result after reporting gas usage and returning the instance into the cache.
    let res = call_handle_raw(&mut instance, params, msg);
    *gas_used = instance.create_gas_report().used_internally;
    instance.recycle();
    Ok(res?)
}

#[no_mangle]
pub extern "C" fn V3_migrate(
    cache: *mut cache_t,
    contract_id: Buffer,
    params: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
    err: Option<&mut Buffer>,
) -> Buffer {
    let r = match to_cache(cache) {
        Some(c) => catch_unwind(AssertUnwindSafe(move || {
            do_migrate(
                c,
                contract_id,
                params,
                msg,
                db,
                api,
                querier,
                gas_limit,
                gas_used,
            )
        }))
        .unwrap_or_else(|_| Err(Error::panic())),
        None => Err(Error::empty_arg(CACHE_ARG)),
    };
    let data = handle_c_error(r, err);
    Buffer::from_vec(data)
}

fn do_migrate(
    cache: &mut Cache<DB, GoApi, GoQuerier>,
    code_id: Buffer,
    params: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
) -> Result<Vec<u8>, Error> {
    let gas_used = gas_used.ok_or_else(|| Error::empty_arg(GAS_USED_ARG))?;
    let code_id: Checksum = unsafe { code_id.read() }
        .ok_or_else(|| Error::empty_arg(CODE_ID_ARG))?
        .try_into()?;
    let params = unsafe { params.read() }.ok_or_else(|| Error::empty_arg(PARAMS_ARG))?;
    let msg = unsafe { msg.read() }.ok_or_else(|| Error::empty_arg(MSG_ARG))?;

    let deps = to_extern(db, api, querier);
    let mut instance = cache.get_instance(&code_id, deps, gas_limit)?;
    // We only check this result after reporting gas usage and returning the instance into the cache.
    let res = call_migrate_raw(&mut instance, params, msg);
    *gas_used = instance.create_gas_report().used_internally;
    instance.recycle();
    Ok(res?)
}

#[no_mangle]
pub extern "C" fn V3_query(
    cache: *mut cache_t,
    code_id: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
    err: Option<&mut Buffer>,
) -> Buffer {
    let r = match to_cache(cache) {
        Some(c) => catch_unwind(AssertUnwindSafe(move || {
            do_query(c, code_id, msg, db, api, querier, gas_limit, gas_used)
        }))
        .unwrap_or_else(|_| Err(Error::panic())),
        None => Err(Error::empty_arg(CACHE_ARG)),
    };
    let data = handle_c_error(r, err);
    Buffer::from_vec(data)
}

fn do_query(
    cache: &mut Cache<DB, GoApi, GoQuerier>,
    code_id: Buffer,
    msg: Buffer,
    db: DB,
    api: GoApi,
    querier: GoQuerier,
    gas_limit: u64,
    gas_used: Option<&mut u64>,
) -> Result<Vec<u8>, Error> {
    let gas_used = gas_used.ok_or_else(|| Error::empty_arg(GAS_USED_ARG))?;
    let code_id: Checksum = unsafe { code_id.read() }
        .ok_or_else(|| Error::empty_arg(CODE_ID_ARG))?
        .try_into()?;
    let msg = unsafe { msg.read() }.ok_or_else(|| Error::empty_arg(MSG_ARG))?;

    let deps = to_extern(db, api, querier);
    let mut instance = cache.get_instance(&code_id, deps, gas_limit)?;
    // We only check this result after reporting gas usage and returning the instance into the cache.
    let res = call_query_raw(&mut instance, msg);
    *gas_used = instance.create_gas_report().used_internally;
    instance.recycle();
    Ok(res?)
}
