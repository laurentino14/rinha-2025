const fs = require("node:fs");
const readline = require("node:readline");

const filePath = process.argv[2]; // pega o primeiro argumento depois do script

if (!filePath) {
	console.error("❌ Uso: node check-duplicate-uuids.js <caminho-do-arquivo>");
	process.exit(1);
}

const seen = new Map();

async function checkDuplicateUUIDs() {
	const fileStream = fs.createReadStream(filePath);
	const rl = readline.createInterface({
		input: fileStream,
		crlfDelay: Infinity,
	});

	let lineNumber = 0;

	for await (const line of rl) {
		lineNumber++;
		if (!line.trim()) continue;

		try {
			const obj = JSON.parse(line);
			const uuid = obj.msg;

			if (!uuid) continue;

			if (!seen.has(uuid)) {
				seen.set(uuid, []);
			}

			seen.get(uuid).push(lineNumber);
		} catch (_e) {
			console.error(`❌ Erro na linha ${lineNumber}: não é JSON válido.`);
		}
	}

	let duplicates = 0;
	for (const [uuid, lines] of seen.entries()) {
		if (lines.length > 1) {
			duplicates++;
			console.log(
				`🔁 UUID duplicado ${uuid} (${lines.length}x), linhas: ${lines.join(", ")}`,
			);
		}
	}

	if (duplicates === 0) {
		console.log("✅ Nenhum UUID duplicado encontrado.");
	}
}

checkDuplicateUUIDs().catch(console.error);
