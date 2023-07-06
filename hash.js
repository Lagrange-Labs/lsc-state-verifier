const { poseidon } = require('@iden3/js-crypto');
const { keccak256, toBuffer, bufferToHex } = require('ethereumjs-util');

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
}

const blsSignatureHash = (bls_signature) => {
    const chain_hash = chainHash(bls_signature.chain_header.block_hash, bls_signature.chain_header.block_number, bls_signature.chain_header.chain_id);

    console.log(chain_hash); // 0x06c3a68be875459bbc887f758c9d4aab01b7eb997daa25b91c99c9fb1e35f14a

    const concated_input = Buffer.concat([toBuffer(chain_hash), toBuffer(bls_signature.current_committee), toBuffer(bls_signature.next_committee), toBuffer(bls_signature.total_voting_power)]);

    return poseidon.hashBytes(concated_input).toString(16);
}

const bls_signature = {
    "chain_header" : {
        "block_hash" : "0x0257554703067ea1fe00f949e542f4a34074f67c457e00a1427d101b823f77fa",
        "block_number" : 13202389,
        "chain_id" : 4321
    },
    "current_committee" : "0x09f582a8133bb26ee103a78a78999466a84455f9a409c46b8599e1aebb95fc8e",
    "next_committee" : "0x22355f09a8afa99cd6c98e0169af50e87d9b7ec858abb60542d4d0139d9aa496",
    "total_voting_power": 10000000
}

console.log(blsSignatureHash(bls_signature)); // d367074837d88aa990d6f191486a36c5f13506772b6f8053c255ce663aab99c