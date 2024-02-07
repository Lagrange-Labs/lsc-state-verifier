"use strict";

const bls = require("@noble/bls12-381");
const { poseidon } = require("@iden3/js-crypto");

const pubkeys = [
  "97f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb",
  "a572cbea904d67468808c8eb50a9450c9721db309128012543902d0ac358a62ae28f75bb8f1c7c42c39a8c5529bf0f4e",
  "89ece308f9d1f0131765212deca99697b112d61f9be9a5f1f3780a51335b3ff981747a0b2ca2179b96d2c0c9024e5224",
  "ac9b60d5afcbd5663a8a44b7c5a02f19e9a77ab0a35bd65809bb5c67ec582c897feb04decc694b13e08587f3ff9b5b60",
  "b0e7791fb972fe014159aa33a98622da3cdc98ff707965e536d8636b5fcc5ac7a91a8c46e59a00dca575af0f18fb13dc",
  "a6e82f6da4520f85c5d27d8f329eccfa05944fd1096b20734c894966d12a9e2a9a9744529d7212d33883113a0cadb909",
  "b928f3beb93519eecf0145da903b40a4c97dca00b21f12ac0df3be9116ef2ef27b2ae6bcd4c5bc2d54ef5a70627efcb7",
  "a85ae765588126f5e860d019c0e26235f567a9c0c0b2d8ff30f3e8d436b1082596e5e7462d20f5be3764fd473e57f9cf",
  "99cdf3807146e68e041314ca93e1fee0991224ec2a74beb2866816fd0826ce7b6263ee31e953a86d1b72cc2215a57793",
  "af81da25ecf1c84b577fefbedd61077a81dc43b00304015b2b596ab67f00e41c86bb00ebd0f90d4b125eb0539891aeed",
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
  "0x40e1201138f0519877e3704F253153C92f5cfD2a",
];

const votingPower = 1000000;
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
  console.log("Merkle root: " + leaves[0].toString(16));
  console.log(
    "\n--------------------------------------------------------------------------------------"
  );
}

testLagrangePubKey();
