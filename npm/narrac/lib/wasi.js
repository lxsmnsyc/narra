"use strict";

// The WASI preview1 calls narrac.wasm imports, written against node:fs. It does not use
// node:wasi. A call that narrac does not import is not here. Run `WebAssembly.Module.imports`
// on the module to see the list.

const fs = require("node:fs");
const path = require("node:path");

const ERRNO = {
  success: 0,
  access: 2,
  badf: 8,
  exist: 20,
  inval: 28,
  io: 29,
  isdir: 31,
  loop: 32,
  nametoolong: 37,
  noent: 44,
  nosys: 52,
  notdir: 54,
  notempty: 55,
  perm: 63,
  xdev: 75,
};

const ERRNO_BY_CODE = {
  EACCES: ERRNO.access,
  EBADF: ERRNO.badf,
  EEXIST: ERRNO.exist,
  EINVAL: ERRNO.inval,
  EISDIR: ERRNO.isdir,
  ELOOP: ERRNO.loop,
  ENAMETOOLONG: ERRNO.nametoolong,
  ENOENT: ERRNO.noent,
  ENOTDIR: ERRNO.notdir,
  ENOTEMPTY: ERRNO.notempty,
  EPERM: ERRNO.perm,
  EXDEV: ERRNO.xdev,
  ENOSPC: 51,
  EMFILE: 33,
  EPIPE: 64,
  EROFS: 69,
};

const FILETYPE = { unknown: 0, characterDevice: 2, directory: 3, regularFile: 4, symbolicLink: 7 };

const RIGHT_READ = 1n << 1n;
const RIGHT_WRITE = 1n << 6n;
const ALL_RIGHTS = 0xffffffffffffffffn;

const OFLAGS = { create: 1, directory: 2, exclusive: 4, truncate: 8 };
const FDFLAGS_APPEND = 1;
const LOOKUP_SYMLINK_FOLLOW = 1;

class ExitSignal {
  constructor(status) {
    this.status = status;
  }
}

class WasiError {
  constructor(errno) {
    this.errno = errno;
  }
}

function findErrno(error) {
  if (error instanceof WasiError) {
    return error.errno;
  }
  if (error && error.code in ERRNO_BY_CODE) {
    return ERRNO_BY_CODE[error.code];
  }
  if (error && error.code === undefined) {
    // A bug in the shim, not a host error: do not hide it as EIO.
    throw error;
  }
  return ERRNO.io;
}

function findFiletype(stats) {
  if (stats.isDirectory()) {
    return FILETYPE.directory;
  }
  if (stats.isFile()) {
    return FILETYPE.regularFile;
  }
  if (stats.isSymbolicLink()) {
    return FILETYPE.symbolicLink;
  }
  if (stats.isCharacterDevice()) {
    return FILETYPE.characterDevice;
  }
  return FILETYPE.unknown;
}

function findOpenFlags(rights, oflags, fdflags) {
  const constants = fs.constants;
  const canRead = (rights & RIGHT_READ) !== 0n;
  const canWrite = (rights & RIGHT_WRITE) !== 0n;
  let flags = constants.O_RDONLY;
  if (canWrite) {
    flags = canRead ? constants.O_RDWR : constants.O_WRONLY;
  }
  if (oflags & OFLAGS.create) {
    flags |= constants.O_CREAT;
  }
  if (oflags & OFLAGS.exclusive) {
    flags |= constants.O_EXCL;
  }
  if (oflags & OFLAGS.truncate) {
    flags |= constants.O_TRUNC;
  }
  if (fdflags & FDFLAGS_APPEND) {
    flags |= constants.O_APPEND;
  }
  return flags;
}

function readStandardInput(buffer) {
  // Loop on EAGAIN (a terminal in non-blocking mode); EOF ends the input.
  for (;;) {
    try {
      return fs.readSync(0, buffer, 0, buffer.length, null);
    } catch (error) {
      if (error.code === "EOF") {
        return 0;
      }
      if (error.code !== "EAGAIN") {
        throw error;
      }
    }
  }
}

function runWasi(module, options) {
  // options: args, env (object), cwd, stdin (Buffer, or undefined for the real fd 0),
  // writeStdout(bytes), writeStderr(bytes). Returns the exit status.
  const preopens = createPreopens(options.cwd, options.args);
  const entries = new Map();
  entries.set(0, { kind: "stdin" });
  entries.set(1, { kind: "stdout" });
  entries.set(2, { kind: "stderr" });
  const preopenNames = new Map();
  for (const preopen of preopens) {
    entries.set(preopen.descriptor, { kind: "directory", path: preopen.hostPath });
    preopenNames.set(preopen.descriptor, preopen.guestName);
  }
  let nextDescriptor = 3 + preopens.length;
  let standardInputOffset = 0;
  let memory;

  const argumentBytes = encodeStrings(options.args);
  const environmentList = [];
  for (const [name, value] of Object.entries(options.env)) {
    if (value !== undefined) {
      environmentList.push(`${name}=${value}`);
    }
  }
  const environmentBytes = encodeStrings(environmentList);

  function viewMemory() {
    return new DataView(memory.buffer);
  }

  function readText(pointer, length) {
    return Buffer.from(memory.buffer, pointer, length).toString("utf8");
  }

  function resolveHostPath(descriptor, pointer, length) {
    // The path is relative to the directory the descriptor names.
    const entry = entries.get(descriptor);
    if (entry === undefined || entry.kind !== "directory") {
      throw new WasiError(ERRNO.badf);
    }
    return path.resolve(entry.path, readText(pointer, length));
  }

  function writeStrings(table, tablePointer, bufferPointer) {
    // table is the list of encoded strings. Each pointer goes to tablePointer, each text to bufferPointer.
    const view = viewMemory();
    const target = new Uint8Array(memory.buffer);
    let offset = 0;
    for (let index = 0; index < table.length; index++) {
      view.setUint32(tablePointer + index * 4, bufferPointer + offset, true);
      target.set(table[index], bufferPointer + offset);
      offset += table[index].length;
    }
  }

  function writeFilestat(pointer, stats) {
    const view = viewMemory();
    view.setBigUint64(pointer, BigInt(stats.dev), true);
    view.setBigUint64(pointer + 8, BigInt(stats.ino), true);
    view.setUint8(pointer + 16, findFiletype(stats));
    view.setBigUint64(pointer + 24, BigInt(stats.nlink), true);
    view.setBigUint64(pointer + 32, BigInt(stats.size), true);
    view.setBigUint64(pointer + 40, BigInt(Math.round(stats.atimeMs * 1e6)), true);
    view.setBigUint64(pointer + 48, BigInt(Math.round(stats.mtimeMs * 1e6)), true);
    view.setBigUint64(pointer + 56, BigInt(Math.round(stats.ctimeMs * 1e6)), true);
  }

  function collectChunks(iovecs, count) {
    const view = viewMemory();
    const chunks = [];
    for (let index = 0; index < count; index++) {
      const pointer = view.getUint32(iovecs + index * 8, true);
      const length = view.getUint32(iovecs + index * 8 + 4, true);
      chunks.push({ pointer, length });
    }
    return chunks;
  }

  // The calls, in the order of the preview1 specification.
  function getArguments(tablePointer, bufferPointer) {
    writeStrings(argumentBytes, tablePointer, bufferPointer);
    return ERRNO.success;
  }

  function getArgumentSizes(countPointer, sizePointer) {
    const view = viewMemory();
    view.setUint32(countPointer, argumentBytes.length, true);
    view.setUint32(sizePointer, sumLengths(argumentBytes), true);
    return ERRNO.success;
  }

  function getEnvironment(tablePointer, bufferPointer) {
    writeStrings(environmentBytes, tablePointer, bufferPointer);
    return ERRNO.success;
  }

  function getEnvironmentSizes(countPointer, sizePointer) {
    const view = viewMemory();
    view.setUint32(countPointer, environmentBytes.length, true);
    view.setUint32(sizePointer, sumLengths(environmentBytes), true);
    return ERRNO.success;
  }

  function getDescriptorStatus(descriptor, pointer) {
    const entry = entries.get(descriptor);
    if (entry === undefined) {
      return ERRNO.badf;
    }
    let filetype = FILETYPE.regularFile;
    if (entry.kind === "directory") {
      filetype = FILETYPE.directory;
    } else if (entry.kind !== "file") {
      filetype = FILETYPE.characterDevice;
    }
    const view = viewMemory();
    view.setUint8(pointer, filetype);
    view.setUint16(pointer + 2, 0, true);
    view.setBigUint64(pointer + 8, ALL_RIGHTS, true);
    view.setBigUint64(pointer + 16, ALL_RIGHTS, true);
    return ERRNO.success;
  }

  function setDescriptorFlags() {
    return ERRNO.success;
  }

  function getPreopen(descriptor, pointer) {
    if (!preopenNames.has(descriptor)) {
      return ERRNO.badf;
    }
    const view = viewMemory();
    view.setUint8(pointer, 0);
    view.setUint32(pointer + 4, Buffer.byteLength(preopenNames.get(descriptor)), true);
    return ERRNO.success;
  }

  function getPreopenName(descriptor, pointer, length) {
    if (!preopenNames.has(descriptor)) {
      return ERRNO.badf;
    }
    const name = Buffer.from(preopenNames.get(descriptor));
    new Uint8Array(memory.buffer).set(name.subarray(0, length), pointer);
    return ERRNO.success;
  }

  function readDescriptor(descriptor, iovecs, count, resultPointer) {
    const entry = entries.get(descriptor);
    if (entry === undefined) {
      return ERRNO.badf;
    }
    let total = 0;
    for (const chunk of collectChunks(iovecs, count)) {
      const target = Buffer.from(memory.buffer, chunk.pointer, chunk.length);
      let got = 0;
      if (entry.kind === "stdin") {
        got = readFromStandardInput(target);
      } else if (entry.kind === "file") {
        got = fs.readSync(entry.hostDescriptor, target, 0, target.length, entry.position);
        entry.position += got;
      } else if (entry.kind === "directory") {
        return ERRNO.isdir;
      } else {
        return ERRNO.badf;
      }
      total += got;
      if (got < chunk.length) {
        break;
      }
    }
    viewMemory().setUint32(resultPointer, total, true);
    return ERRNO.success;
  }

  function readFromStandardInput(target) {
    if (options.stdin === undefined) {
      return readStandardInput(target);
    }
    const rest = options.stdin.subarray(standardInputOffset, standardInputOffset + target.length);
    target.set(rest);
    standardInputOffset += rest.length;
    return rest.length;
  }

  function writeDescriptor(descriptor, iovecs, count, resultPointer) {
    const entry = entries.get(descriptor);
    if (entry === undefined) {
      return ERRNO.badf;
    }
    let total = 0;
    for (const chunk of collectChunks(iovecs, count)) {
      // Copy out of linear memory: the host call may outlive a memory growth.
      const bytes = Buffer.from(Buffer.from(memory.buffer, chunk.pointer, chunk.length));
      if (entry.kind === "stdout") {
        options.writeStdout(bytes);
      } else if (entry.kind === "stderr") {
        options.writeStderr(bytes);
      } else if (entry.kind === "file") {
        writeFile(entry, bytes);
      } else {
        return ERRNO.badf;
      }
      total += bytes.length;
    }
    viewMemory().setUint32(resultPointer, total, true);
    return ERRNO.success;
  }

  function writeFile(entry, bytes) {
    let offset = 0;
    while (offset < bytes.length) {
      const position = entry.append ? null : entry.position;
      const written = fs.writeSync(
        entry.hostDescriptor,
        bytes,
        offset,
        bytes.length - offset,
        position,
      );
      offset += written;
      entry.position += written;
    }
  }

  function seekDescriptor(descriptor, offset, whence, resultPointer) {
    const entry = entries.get(descriptor);
    if (entry === undefined || entry.kind !== "file") {
      return entry === undefined ? ERRNO.badf : ERRNO.inval;
    }
    let base = 0;
    if (whence === 1) {
      base = entry.position;
    } else if (whence === 2) {
      base = fs.fstatSync(entry.hostDescriptor).size;
    }
    const target = base + Number(offset);
    if (target < 0) {
      return ERRNO.inval;
    }
    entry.position = target;
    viewMemory().setBigUint64(resultPointer, BigInt(target), true);
    return ERRNO.success;
  }

  function closeDescriptor(descriptor) {
    const entry = entries.get(descriptor);
    if (entry === undefined) {
      return ERRNO.badf;
    }
    if (entry.kind === "file") {
      fs.closeSync(entry.hostDescriptor);
    }
    entries.delete(descriptor);
    return ERRNO.success;
  }

  function readDirectory(descriptor, bufferPointer, bufferLength, cookie, resultPointer) {
    const entry = entries.get(descriptor);
    if (entry === undefined || entry.kind !== "directory") {
      return entry === undefined ? ERRNO.badf : ERRNO.notdir;
    }
    const names = fs.readdirSync(entry.path, { withFileTypes: true });
    const target = new Uint8Array(memory.buffer, bufferPointer, bufferLength);
    const header = new DataView(new ArrayBuffer(24));
    let used = 0;
    for (let index = Number(cookie); index < names.length && used < bufferLength; index++) {
      const name = Buffer.from(names[index].name);
      header.setBigUint64(0, BigInt(index + 1), true);
      header.setBigUint64(8, 0n, true);
      header.setUint32(16, name.length, true);
      header.setUint8(20, findFiletype(names[index]));
      // A last entry that does not fit is cut off; the caller sees a full buffer and reads again.
      const record = Buffer.concat([Buffer.from(header.buffer), name]);
      const slice = record.subarray(0, bufferLength - used);
      target.set(slice, used);
      used += slice.length;
    }
    viewMemory().setUint32(resultPointer, used, true);
    return ERRNO.success;
  }

  function getPathStatus(descriptor, flags, pathPointer, pathLength, resultPointer) {
    const hostPath = resolveHostPath(descriptor, pathPointer, pathLength);
    if (
      findVirtualDevice(descriptor, readText(pathPointer, pathLength)) !== undefined &&
      !fs.existsSync(hostPath)
    ) {
      // A stream has no size or times: report a character device.
      new Uint8Array(memory.buffer, resultPointer, 64).fill(0);
      viewMemory().setUint8(resultPointer + 16, FILETYPE.characterDevice);
      return ERRNO.success;
    }
    const stats = flags & LOOKUP_SYMLINK_FOLLOW ? fs.statSync(hostPath) : fs.lstatSync(hostPath);
    writeFilestat(resultPointer, stats);
    return ERRNO.success;
  }

  function openPath(
    descriptor,
    dirflags,
    pathPointer,
    pathLength,
    oflags,
    rightsBase,
    rightsInheriting,
    fdflags,
    resultPointer,
  ) {
    const hostPath = resolveHostPath(descriptor, pathPointer, pathLength);
    const device = findVirtualDevice(descriptor, readText(pathPointer, pathLength));
    if (device !== undefined && !fs.existsSync(hostPath)) {
      // A stream of the call: a new descriptor that reads or writes it.
      const aliased = nextDescriptor++;
      entries.set(aliased, { kind: device });
      viewMemory().setUint32(resultPointer, aliased, true);
      return ERRNO.success;
    }
    let stats;
    try {
      stats = dirflags & LOOKUP_SYMLINK_FOLLOW ? fs.statSync(hostPath) : fs.lstatSync(hostPath);
    } catch (error) {
      if (error.code !== "ENOENT") {
        throw error;
      }
    }
    if (stats !== undefined && oflags & OFLAGS.create && oflags & OFLAGS.exclusive) {
      return ERRNO.exist;
    }
    if (oflags & OFLAGS.directory && stats !== undefined && !stats.isDirectory()) {
      return ERRNO.notdir;
    }
    let entry;
    if (stats !== undefined && stats.isDirectory()) {
      entry = { kind: "directory", path: hostPath };
    } else {
      const flags = findOpenFlags(rightsBase, oflags, fdflags);
      const hostDescriptor = fs.openSync(hostPath, flags, 0o666);
      entry = {
        kind: "file",
        hostDescriptor,
        position: 0,
        append: (fdflags & FDFLAGS_APPEND) !== 0,
      };
    }
    const opened = nextDescriptor++;
    entries.set(opened, entry);
    viewMemory().setUint32(resultPointer, opened, true);
    return ERRNO.success;
  }

  function removeDirectory(descriptor, pathPointer, pathLength) {
    fs.rmdirSync(resolveHostPath(descriptor, pathPointer, pathLength));
    return ERRNO.success;
  }

  function unlinkFile(descriptor, pathPointer, pathLength) {
    fs.unlinkSync(resolveHostPath(descriptor, pathPointer, pathLength));
    return ERRNO.success;
  }

  function exitProcess(status) {
    throw new ExitSignal(status);
  }

  const calls = {
    args_get: getArguments,
    args_sizes_get: getArgumentSizes,
    environ_get: getEnvironment,
    environ_sizes_get: getEnvironmentSizes,
    fd_fdstat_get: getDescriptorStatus,
    fd_fdstat_set_flags: setDescriptorFlags,
    fd_prestat_get: getPreopen,
    fd_prestat_dir_name: getPreopenName,
    fd_read: readDescriptor,
    fd_readdir: readDirectory,
    fd_seek: seekDescriptor,
    fd_write: writeDescriptor,
    fd_close: closeDescriptor,
    path_filestat_get: getPathStatus,
    path_open: openPath,
    path_remove_directory: removeDirectory,
    path_unlink_file: unlinkFile,
    proc_exit: exitProcess,
  };

  // Wrap each call: a host error becomes an errno, and proc_exit passes through.
  const imports = {};
  for (const [name, call] of Object.entries(calls)) {
    imports[name] = function callWasi(...callArguments) {
      try {
        return call(...callArguments);
      } catch (error) {
        if (error instanceof ExitSignal) {
          throw error;
        }
        return findErrno(error);
      }
    };
  }
  for (const wanted of WebAssembly.Module.imports(module)) {
    if (wanted.module === "wasi_snapshot_preview1" && !(wanted.name in imports)) {
      imports[wanted.name] = function failWasi() {
        return ERRNO.nosys;
      };
    }
  }

  const instance = new WebAssembly.Instance(module, { wasi_snapshot_preview1: imports });
  memory = instance.exports.memory;
  try {
    instance.exports._start();
    return 0;
  } catch (error) {
    if (error instanceof ExitSignal) {
      return error.status;
    }
    throw error;
  } finally {
    closeOpenFiles(entries);
  }
}

function closeOpenFiles(entries) {
  for (const entry of entries.values()) {
    if (entry.kind === "file") {
      fs.closeSync(entry.hostDescriptor);
    }
  }
}

function createPreopens(cwd, args) {
  // wasi-libc strips the leading "/" and "./" from a preopen name and from every path, so "." and
  // "/" are the same empty prefix, and an absolute path looks like a relative one. The empty prefix
  // therefore names the working directory (relative paths, `..` included). An absolute path is
  // matched by the longest prefix: the top folder of each absolute argument gets a preopen of its
  // own, which the empty prefix loses to.
  const preopens = [{ descriptor: 3, guestName: ".", hostPath: path.resolve(cwd) }];
  const seen = new Set();
  for (const argument of args) {
    if (!argument.startsWith("/")) {
      continue;
    }
    const top = argument.split("/")[1];
    if (top !== "" && top !== undefined && !seen.has(top)) {
      seen.add(top);
      preopens.push({ descriptor: 3 + preopens.length, guestName: `/${top}`, hostPath: `/${top}` });
    }
  }
  return preopens;
}

function findVirtualDevice(descriptor, text) {
  // The compiler reads "/dev/stdin" as a file. Under the working directory it is the stream of
  // the call, unless the folder has a dev/stdin.
  const names = { "dev/stdin": "stdin", "dev/stdout": "stdout", "dev/stderr": "stderr" };
  const kind = names[text.replace(/^\/+/, "")];
  return descriptor === 3 && kind !== undefined ? kind : undefined;
}

function encodeStrings(texts) {
  const encoded = [];
  for (const text of texts) {
    encoded.push(Buffer.from(`${text}\0`));
  }
  return encoded;
}

function sumLengths(buffers) {
  let total = 0;
  for (const buffer of buffers) {
    total += buffer.length;
  }
  return total;
}

module.exports = { runWasi };
