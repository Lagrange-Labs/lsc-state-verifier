const { poseidon } = require("@iden3/js-crypto");
const { keccak256, toBuffer, bufferToHex } = require("ethereumjs-util");

// Define a function to compute the Keccak hash of elements
function Hash(...data) {
	// Concatenate the data elements
	const concatenatedData = Buffer.concat(data);

	// Compute the Keccak-256 hash
	const hash = keccak256(concatenatedData);

	// Return the hash as a Uint8Array
	return hash;
}

const chainHash = (blockHash, blockNumber, chainID) => {
	const blockNumberBuf = Buffer.alloc(32);
	const chainIDBuf = Buffer.alloc(4);

	blockNumberBuf.writeBigUInt64BE(BigInt(blockNumber), 24);
	chainIDBuf.writeUInt32BE(chainID, 0);

	return bufferToHex(Hash(toBuffer(blockHash), blockNumberBuf, chainIDBuf));
};

const blsSignatureHash = (bls_signature) => {
	const chain_hash = chainHash(
		bls_signature.chain_header.block_hash,
		bls_signature.chain_header.block_number,
		bls_signature.chain_header.chain_id
	);

	console.log("Chain header keccak hash: " + chain_hash);

	const concated_input = Buffer.concat([
		toBuffer(chain_hash),
		toBuffer(bls_signature.current_committee),
		toBuffer(bls_signature.next_committee),
		toBuffer(bls_signature.total_voting_power),
	]);

	return poseidon.hashBytes(concated_input).toString(16);
};

const bls_signature = {
	chain_header: {
		block_hash:
			"0x95aea085c0d4a908eed989c9f2c793477d53309ae3e9f0a28f29510ffeff2b91",
		block_number: 28810640,
		chain_id: 421613,
	},
	current_committee:
		"0x2e3d2e5c97ee5320cccfd50434daeab6b0072558b693bb0e7f2eeca97741e514",
	next_committee:
		"0x2e3d2e5c97ee5320cccfd50434daeab6b0072558b693bb0e7f2eeca97741e514",
	total_voting_power: 900000000,
};

console.log("Signing root: " + blsSignatureHash(bls_signature));
