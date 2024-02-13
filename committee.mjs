"use strict";

import { PointG1 } from "./index.js";
import pkg from "@iden3/js-crypto";
const { poseidon } = pkg;

const X_API_KEY = ""; // TODO: add API key
const CHAIN_CONFIG = {
  OPTIMISM: "420",
};
const BLOCK_NUMBER = 19635965;
const URI =
  "http://lagrange-query-layer-env.eba-ygn6m8ig.us-east-1.elasticbeanstalk.com";
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

async function getCommitteeRoot(pubkeys, address, votingPower) {
  for (let i = 0; i < pubkeys.length; i++) {
    let publicKey = PointG1.fromHex(pubkeys[i]);
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

    const votingPowerStr = votingPower[i].toString(16).padStart(32, "0");
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

  return leaves[0].toString(16);
}

async function fetchBlockData(block_number, chain_id) {
  try {
    const response = await fetch(
      `${URI}/blocks/block-data?block_number=${block_number}&chain_id=${chain_id}`,
      {
        headers: {
          "x-api-key": X_API_KEY,
        },
      }
    );
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    } else {
      const data = await response.json();
      // console.log(data);
      return data;
    }
  } catch (error) {
    console.log(
      "There was a problem with the fetch operation: " + error.message
    );
  }
}

async function generateCommitteeRoot(block_number, chain_id) {
  let data = await fetchBlockData(block_number, chain_id);
  let committeeRoot = getCommitteeRoot(
    data.pubkeys,
    data.address,
    data.votingPower
  );
  return committeeRoot;
}

generateCommitteeRoot(BLOCK_NUMBER, CHAIN_CONFIG.OPTIMISM)
  .then((data) => console.log("Committee root generated successfully"))
  .catch((error) => console.error("Error:", error));
