const { poseidon } = require('@iden3/js-crypto');
const { keccak256, toBuffer, bufferToHex } = require('ethereumjs-util');
const {bls12_381} = require('@noble/curves/bls12-381');

const keys = require('./keys.json');

// Define a function to compute the Keccak hash of elements
function keccakHash(...data) {
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

    return bufferToHex(keccakHash(toBuffer(blockHash), blockNumberBuf, chainIDBuf));
}

const blsSignatureHash = (block) => {
    const chain_hash = chainHash(block.chain_header.block_hash, block.chain_header.block_number, block.chain_header.chain_id);

    console.log("Chain header keccak hash: ", chain_hash);

    const concated_input = Buffer.concat([toBuffer(chain_hash), toBuffer("0x" + block.block_header.current_committee), toBuffer("0x" + block.block_header.next_committee)]);

    bn_hash = poseidon.hashBytes(concated_input).toString(16);
    padded_hash = bn_hash.length % 2 == 0 ? bn_hash : "0" + bn_hash;
    return padded_hash;
}

const verifyBLS = async (block) => {
    const agg_pub_key = bls12_381.aggregatePublicKeys(block.pub_keys);
    const msg = blsSignatureHash(block);

    const sigs = [];
    for (const pub_key of block.pub_keys) {
        const priv_key = keys['0x' + pub_key].slice(2);
        console.log("priv key: ", priv_key);
        const sig = await bls12_381.sign(msg, priv_key);
        console.log("bls pub key: ", Buffer.from(bls12_381.getPublicKey(priv_key)).toString('hex'));
        console.log(pub_key, "Signature: ", Buffer.from(sig).toString('hex'));
        sigs.push(sig);
    }

    const agg_sig = bls12_381.aggregateSignatures(sigs);
    console.log("Aggregate signature: ", Buffer.from(agg_sig).toString('hex'));

    console.log("Signing root (message): ", msg);
    console.log("Aggregate public key: ", agg_pub_key);
    pg = bls12_381.PointG1.fromHex(agg_pub_key);
    console.log(pg.toAffine().map(x => x.value.toString(16)));
    const resJS = await bls12_381.verify(agg_sig, msg, agg_pub_key);
    console.log("BLS signature verification result: ", resJS);
    const res = await bls12_381.verify(block.agg_signature, msg, agg_pub_key);
    console.log("BLS signature verification result: ", res);

}

const block = {
    "_id": "649e027f951f3067cfef4d0e",
    "agg_signature": "a1642c2990459092b00cf76d481969ce27cf2edccd606969b835f038a818f27be500b2aa1d5ed6b068df4b99d0094a8f09963550e207e46c77d5630121c4ce4597e4621eff696b308801e728a5c0bf9e306a14d39c7dde014fd78b87ecf53690",
    "block_header": {
        "current_committee": "2e3d2e5c97ee5320cccfd50434daeab6b0072558b693bb0e7f2eeca97741e514",
        "epoch_block_number": 9263590,
        "next_committee": "2e3d2e5c97ee5320cccfd50434daeab6b0072558b693bb0e7f2eeca97741e514",
        "proposer_pub_key": "86b50179774296419b7e8375118823ddb06940d9a28ea045ab418c7ecbe6da84d416cb55406eec6393db97ac26e38bd4",
        "proposer_signature": "a4a9d0dc9550a44ff6f7b46d2ab97d2e84754e0c2bf3be7e2aa4a824a934d75439fdefb59aa947f34ce4b36c71fa78600b578577e8b5ef54ddc97067297dfccebc7dec27e9f6adc98825dabb3a096c7406c0cba626abd2d787664273636ac32e",
        "total_voting_power": 900000000
    },
    "chain_header": {
        "block_hash": "0x95aea085c0d4a908eed989c9f2c793477d53309ae3e9f0a28f29510ffeff2b91",
        "block_number": 28810640,
        "chain_id": 421613
    },
    "pub_keys": [
        "86b50179774296419b7e8375118823ddb06940d9a28ea045ab418c7ecbe6da84d416cb55406eec6393db97ac26e38bd4",
        "a771bcc2d948fa9801916e0c18aa27732f77d43b751138a3e119928103826427a5c11080620f9a9e37d8f4de3a4868cf",
        "91f8a57e971b831b80f1f6c04eed120c2c3ecb9d13adf7601e05614909158c3882bdbc04251cb689ecf653c28b981774",
        "9496160c06e86aae28d251e082173b49ff10225f29611fae766d1a1a473d4bcf1188738208df163c01927b8c5df5160a",
        "837ca7f100239253d16d982102e47387a369cd1ba4bf9cf08ab85bd70c2b7559e7158196d9c60d03a571dfb3580b2b8e",
        "b5695acd75a5d52e82eddf4ae1c01a1d456085da4ce255169cdac877b5a622386e7db6cb9c4c39b4eaf660dfa3a80d5d",
        "90aa5215784d04f54d2df41cca557514c25b26cd0803da59c08eb3954633c4fe5131e195b1483123f5c9fe9f08200a6d"
    ]
}

verifyBLS(block);