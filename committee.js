"use strict";

const bls = require('@noble/bls12-381');
const { poseidon } = require('@iden3/js-crypto');

const pubkeys = [
	"86b50179774296419b7e8375118823ddb06940d9a28ea045ab418c7ecbe6da84d416cb55406eec6393db97ac26e38bd4",
	"90aa5215784d04f54d2df41cca557514c25b26cd0803da59c08eb3954633c4fe5131e195b1483123f5c9fe9f08200a6d",
	"a771bcc2d948fa9801916e0c18aa27732f77d43b751138a3e119928103826427a5c11080620f9a9e37d8f4de3a4868cf",
	"b5695acd75a5d52e82eddf4ae1c01a1d456085da4ce255169cdac877b5a622386e7db6cb9c4c39b4eaf660dfa3a80d5d",
	"9496160c06e86aae28d251e082173b49ff10225f29611fae766d1a1a473d4bcf1188738208df163c01927b8c5df5160a",
	"837ca7f100239253d16d982102e47387a369cd1ba4bf9cf08ab85bd70c2b7559e7158196d9c60d03a571dfb3580b2b8e",
	"91f8a57e971b831b80f1f6c04eed120c2c3ecb9d13adf7601e05614909158c3882bdbc04251cb689ecf653c28b981774",
	"b004ff3a1e4aa11af5d951d7aedc33ecca16a6bb7e639e81650659c6f896cddcaa82bd8e33c0b991776140731129b1e5",
	"969d9fc322c302e418d6e73367bfb001d5a62008095126e10672f355918398fc3a70b0a96ec262693e9c6de301cd10fe",
];

const address = [
	"0x6E654b122377EA7f592bf3FD5bcdE9e8c1B1cEb9",
	"0x516D6C27C23CEd21BF7930E2a01F0BcA9A141a0d",
	"0x4d694DE17246086d6451D732Ea8EA2a9a76dC997",
	"0x5d51B4c1fb0c67d0e1274EC96c1B895F45505a3D",
	"0x13cF11F76a08214A826355a1C8d661E41EA7Bf97",
	"0xBD2369a9535751004617bC47cB0BF8Ea5c35Ed7C",
	"0x83070c799c0d41526D4c71e462557CdbB2C750AC",
	"0x7365666466f97E8aBBEE8900925521e0469A1f25",
	"0xaa58F0fC9eddeFBef9E4C1c3e81f8cA5f22b9B8e",
];

const votingPower = 100000000;
var leaves = [];

function bigint_to_array(n, k, x) {
	let mod = 1n;
	for (var idx = 0; idx < n; idx++) {
		mod = mod * 2n;
	}
	let ret = [];
	var x_temp = x;
	for (var idx = 0; idx < k; idx++) {
		ret.push((x_temp % mod).toString());
		x_temp = x_temp / mod;
	}
	return ret;
}
const hexes = Array.from({ length: 256 }, (v, i) =>
	i.toString(16).padStart(2, "0")
);
function bytesToHex(uint8a) {
	let hex = "";
	for (let i = 0; i < uint8a.length; i++) {
		hex += hexes[uint8a[i]];
	}
	return hex;
}

function bytesToNumberBE(uint8a) {
	if (!(uint8a instanceof Uint8Array)) throw new Error("Expected Uint8Array");
	return BigInt("0x" + bytesToHex(Uint8Array.from(uint8a)));
}

function bigIntToHex(bigInt) {
	let hex = bigInt.toString(16); // convert to hexadecimal string
	if (hex.length % 2 !== 0) {
		hex = "0" + hex;
	}
	return hex;
}

async function testLagrangePubKey() {
	for (let i = 0; i < pubkeys.length; i++) {
		let publicKey = bls.PointG1.fromHex(pubkeys[i]);
		const Gx = publicKey.toAffine()[0].value.toString(16).padStart(96, "0");
		const Gy = publicKey.toAffine()[1].value.toString(16).padStart(96, "0");
		console.log("\npublicKey-" + i + ": " + pubkeys[i] + "\n");
		console.log("Gx = " + Gx);
		console.log("Gy = " + Gy);
		console.log("\n");
		console.log(
			JSON.stringify([
				bigint_to_array(55, 7, publicKey.toAffine()[0].value),
				bigint_to_array(55, 7, publicKey.toAffine()[1].value),
			])
		);

		const chunks = [];
		for (let j = 0; j < 4; j++) {
			chunks.push(BigInt("0x" + Gx.slice(j * 24, (j + 1) * 24)));
		}

		for (let j = 0; j < 4; j++) {
			chunks.push(BigInt("0x" + Gy.slice(j * 24, (j + 1) * 24)));
		}

		const votingPowerStr = votingPower.toString(16).padStart(32, "0");
		chunks.push(BigInt(address[i].slice(0, 26)));
		chunks.push(
			BigInt("0x" + address[i].slice(26, 42) + votingPowerStr.slice(0, 8))
		);
		chunks.push(BigInt("0x" + votingPowerStr.slice(8, 32)));

		const left = poseidon.hash(chunks.slice(0, 6));
		const right = poseidon.hash(chunks.slice(6, 11));
		const leaf = poseidon.hash([left, right]);
		console.log("Chunks: ");
		console.log(
			chunks.map((e) => {
				return e.toString(16);
			})
		);
		console.log("\nLeaf hash: " + leaf.toString(16));
		console.log(
			"\n--------------------------------------------------------------------------------------"
		);

		leaves.push(leaf.toString(16));
	}
	console.log("\nLeaves: \n");
	for (let i = 0; i < leaves.length; i++) {
		console.log(leaves[i]);
		leaves[i] = BigInt("0x" + leaves[i]);
	}
	console.log("\n");

	let count = 1;
	while (count < leaves.length) {
		count *= 2;
	}
	const len_without_zeros = leaves.length;
	for (let i = 0; i < count - len_without_zeros; i++) {
		leaves.push(BigInt(0));
	}

	console.log("\nLeaves with padding: " + leaves.length);
	for (let i = 0; i < leaves.length; i++) {
		console.log(leaves[i]);
	}

	while (count > 1) {
		for (let i = 0; i < count; i += 2) {
			const left = leaves[i];
			const right = leaves[i + 1];
			console.log("\nleft: " + left + "\nright: " + right);
			const hash = poseidon.hash([left, right]);
			console.log("hash: " + hash);
			leaves[i / 2] = hash;
		}
		count /= 2;
	}

	console.log(
		"\n--------------------------------------------------------------------------------------"
	);
	console.log("Merkle root: " + leaves[0].toString(16)); // 2e3d2e5c97ee5320cccfd50434daeab6b0072558b693bb0e7f2eeca97741e514
	console.log(
		"\n--------------------------------------------------------------------------------------"
	);
}

testLagrangePubKey();
