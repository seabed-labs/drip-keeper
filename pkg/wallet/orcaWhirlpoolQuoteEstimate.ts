import NodeWallet from "@project-serum/anchor/dist/cjs/nodewallet";
import {AnchorProvider, BN} from "@project-serum/anchor";
import {
    AccountFetcher,
    buildWhirlpoolClient,
    ORCA_WHIRLPOOL_PROGRAM_ID,
    PDAUtil, swapQuoteByInputToken,
    WhirlpoolContext
} from "@orca-so/whirlpools-sdk";
import {Percentage} from "@orca-so/common-sdk";
import { PublicKey, Keypair, Connection } from "@solana/web3.js";

async function getQuote() {
    const args = process.argv.slice(2);
    if (args.length < 5) {
        console.log(JSON.stringify({
            error: `invalid number of arguments ${args.length}`
        }));
        return;
    }
    const config = new PublicKey(args[0]);
    const tokenAMint = new PublicKey(args[1]);
    const tokenBMint = new PublicKey(args[2]);
    const inputToken = new PublicKey(args[3]);
    const connection = args[4] === "MAINNET"? "https://api.mainnet-beta.solana.com" : "https://api.devnet.solana.com";

    // Don't need to sign anything, so a random keypair is fine
    const wallet = new NodeWallet(Keypair.generate());
    const provider = new AnchorProvider(
        new Connection(connection, "confirmed"),
        wallet,
        AnchorProvider.defaultOptions()
    );
    const fetcher = new AccountFetcher(provider.connection);
    // @ts-ignore - orca uses an older anchor version, so the provider is incompatible
    const ctx = WhirlpoolContext.withProvider(provider, ORCA_WHIRLPOOL_PROGRAM_ID);
    const whirlpoolClient = buildWhirlpoolClient(ctx);

    const whirlpoolPda = PDAUtil.getWhirlpool(
        ORCA_WHIRLPOOL_PROGRAM_ID,
        config,
        tokenAMint,
        tokenBMint,
        64,
    );

    const whirlpool = await whirlpoolClient.getPool(whirlpoolPda.publicKey, true);

    const swapQuote =  await swapQuoteByInputToken(
        whirlpool,
        inputToken,
        new BN(100),
        Percentage.fromFraction(0, 100),
        ORCA_WHIRLPOOL_PROGRAM_ID,
        fetcher,
        true,
    );
    const swapQuoteString =  {
        estimatedAmountIn: swapQuote.estimatedAmountIn.toString(),
        estimatedAmountOut: swapQuote.estimatedAmountOut.toString(),
        estimatedEndTickIndex: swapQuote.estimatedEndTickIndex,
        estimatedEndSqrtPrice: swapQuote.estimatedEndSqrtPrice.toString(),
        estimatedFeeAmount: swapQuote.estimatedFeeAmount.toString(),
        amount: swapQuote.amount.toString(),
        amountSpecifiedIsInput: swapQuote.amountSpecifiedIsInput,
        aToB: swapQuote.aToB,
        otherAmountThreshold: swapQuote.otherAmountThreshold.toString(),
        sqrtPriceLimit: swapQuote.sqrtPriceLimit.toString(),
        tickArray0: swapQuote.tickArray0.toString(),
        tickArray1: swapQuote.tickArray1.toString(),
        tickArray2: swapQuote.tickArray2.toString(),
    };
    console.log(JSON.stringify(swapQuoteString));
}

try {
    getQuote();
} catch(e) {
    console.log(JSON.stringify({
        error: e
    }));
}
