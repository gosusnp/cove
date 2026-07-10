// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Creates a ZIP archive of the mobile bundle for OTA updates.
// Usage: node scripts/zip-mobile.mjs <src-dir> <out-file>

import { ZipArchive } from "archiver";
import { createWriteStream } from "node:fs";

const srcDir = process.argv[2];
const outFile = process.argv[3];
if (!srcDir || !outFile) {
	console.error("Usage: node scripts/zip-mobile.mjs <src-dir> <out-file>");
	process.exit(1);
}

const output = createWriteStream(outFile);
const archive = new ZipArchive({ zlib: { level: 9 } });

archive.on("error", (err) => {
	throw err;
});

archive.pipe(output);
archive.directory(srcDir, false);
await archive.finalize();
console.log(`zipped ${archive.pointer()} bytes → ${outFile}`);
