import { SolUtils } from "./solana-programs/tests/utils/SolUtils";
import { TokenUtil } from "./solana-programs/tests/utils/Token.util";
import {
  findAssociatedTokenAddress,
  generatePairs,
} from "./solana-programs/tests/utils/common.util";
import {
  deploySwap,
  deployVault,
  deployVaultPeriod,
  deployVaultProtoConfig,
  depositWithNewUserWrapper,
  sleep,
} from "./solana-programs/tests/utils/setup.util";
import YAML from "yaml";
import * as fs from "fs";

export async function setupKeeperBot() {
  // let tokenOwnerKeypair: Keypair;
  // let payerKeypair: Keypair;

  // let tokenA: Token;
  // let tokenB: Token;
  // let swap: PublicKey;
  // let vaultProtoConfig: PublicKey;
  // let vaultPDA: PDA;
  // let vaultPeriods: PDA[];
  // let vaultTokenA_ATA: PublicKey;
  // let vaultTokenB_ATA: PublicKey;

  // let swapTokenMint: PublicKey;
  // let swapTokenAAccount: PublicKey;
  // let swapTokenBAccount: PublicKey;
  // let swapFeeAccount: PublicKey;
  // let swapAuthority: PublicKey;

  // let depositWithNewUser;

  // https://discord.com/channels/889577356681945098/889702325231427584/910244405443715092
  // sleep to progress to the next block
  await sleep(500);

  const [tokenOwnerKeypair, payerKeypair] = generatePairs(2);
  await Promise.all([
    SolUtils.fundAccount(payerKeypair.publicKey, 1000000000),
    SolUtils.fundAccount(tokenOwnerKeypair.publicKey, 1000000000),
  ]);

  console.log("tokenOwnerKeypair:", {
    publicKey: tokenOwnerKeypair.publicKey.toString(),
    secretKey: tokenOwnerKeypair.secretKey.toString(),
  });

  console.log("payerKeypair:", {
    publicKey: payerKeypair.publicKey.toString(),
    secretKey: payerKeypair.secretKey.toString(),
  });

  const tokenA = await TokenUtil.createMint(
    tokenOwnerKeypair.publicKey,
    null,
    6,
    payerKeypair
  );
  console.log("tokenAMint:", tokenA.publicKey.toBase58());

  const tokenB = await TokenUtil.createMint(
    tokenOwnerKeypair.publicKey,
    null,
    6,
    payerKeypair
  );
  console.log("tokenBMint:", tokenB.publicKey.toBase58());

  const [
    swap,
    swapTokenMint,
    swapTokenAAccount,
    swapTokenBAccount,
    swapFeeAccount,
    swapAuthority,
  ] = await deploySwap(
    tokenA,
    tokenOwnerKeypair,
    tokenB,
    tokenOwnerKeypair,
    payerKeypair
  );
  console.log("swap:", swap.toBase58());
  console.log("swapTokenMint:", swapTokenMint.toBase58());
  console.log("swapTokenAAccount:", swapTokenAAccount.toBase58());
  console.log("swapTokenBAccount:", swapTokenBAccount.toBase58());
  console.log("swapFeeAccount:", swapFeeAccount.toBase58());
  console.log("swapAuthority:", swapAuthority.toBase58());

  const vaultProtoConfig = await deployVaultProtoConfig(1);
  console.log("vaultProtoConfig:", vaultProtoConfig.toBase58());

  const vaultPDA = await deployVault(
    tokenA.publicKey,
    tokenB.publicKey,
    vaultProtoConfig
  );
  console.log("vault:", vaultPDA.publicKey.toBase58());

  const [vaultTokenA_ATA, vaultTokenB_ATA] = await Promise.all([
    findAssociatedTokenAddress(vaultPDA.publicKey, tokenA.publicKey),
    findAssociatedTokenAddress(vaultPDA.publicKey, tokenB.publicKey),
  ]);
  console.log("vaultTokenAAccount:", vaultTokenA_ATA.toBase58());
  console.log("vaultTokenBAccount:", vaultTokenB_ATA.toBase58());

  const numPeriods = 101;
  const vaultPeriods = await Promise.all(
    [...Array(numPeriods).keys()].map((i) =>
      deployVaultPeriod(
        vaultProtoConfig,
        vaultPDA.publicKey,
        tokenA.publicKey,
        tokenB.publicKey,
        i
      )
    )
  );
  console.log(`deployed ${numPeriods} vault periods`);

  const depositWithNewUser = depositWithNewUserWrapper(
    vaultPDA.publicKey,
    tokenOwnerKeypair,
    tokenA
  );

  for (let i = 1; i < 11; i++) {
    await depositWithNewUser({
      dcaCycles: i * 10,
      newUserEndVaultPeriod: vaultPeriods[i * 10].publicKey,
      mintAmount: i,
    });
  }

  const localConfig = {
    environment: "LOCALNET",
    vault: vaultPDA.publicKey.toBase58(),
    vaultProtoConfig: vaultProtoConfig.toBase58(),
    vaultTokenAAccount: vaultTokenA_ATA.toBase58(),
    vaultTokenBAccount: vaultTokenB_ATA.toBase58(),
    tokenAMint: tokenA.publicKey.toBase58(),
    tokenBMint: tokenB.publicKey.toBase58(),
    swapTokenMint: swapTokenMint.toBase58(),
    swapTokenAAccount: swapTokenAAccount.toBase58(),
    swapTokenBAccount: swapTokenBAccount.toBase58(),
    swapFeeAccount: swapFeeAccount.toBase58(),
    swapAuthority: swapAuthority.toBase58(),
    swap: swap.toBase58(),
  };
  fs.writeFileSync("../configs/localnet1.yaml", YAML.stringify(localConfig));
}

if (require.main === module) {
  setupKeeperBot();
}
