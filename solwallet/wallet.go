package solwallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/pkg/bincode"
	"github.com/blocto/solana-go-sdk/program/associated_token_account"
	"github.com/blocto/solana-go-sdk/program/compute_budget"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/types"
	"log"
)

type Instruction uint8

// 参考 https://github.com/raydium-io/raydium-amm/blob/master/program/src/instruction.rs#L42
const (
	InstructionInitializet Instruction = iota
	InstructionInitialize2
	InstructionMonitorStep
	InstructionDeposit
	InstructionWithdraw
	InstructionMigrateToOpenBook
	InstructionSetParams
	InstructionWithdrawPnl
	InstructionWithdrawSrm
	InstructionSwapBaseIn
	InstructionPreInitialize
	InstructionSwapBaseOut
	InstructionSimulateInfo
	InstructionAdminCancelOrders
	InstructionCreateConfigAccount
	InstructionUpdateConfigAccount
	DEV = "https://solana-devnet.g.alchemy.com/v2/wqZxT7UnY6AgrzV42CtGgGQ7ZGM-UrTq"
	//Piv = "46M2pAp4z3mNPuTh7jS8XHSn69TC4FAnX33Avjx7wqy3W1zzZKiSoBmNTH5PEBDKu7xR2rPa9ocSyzGWYFK7VRF2"
)

// SendSol 用于组装sol币的的交易
func SendSol() {
	// 拆解
	DEV := "https://solana-devnet.g.alchemy.com/v2/wqZxT7UnY6AgrzV42CtGgGQ7ZGM-UrTq"
	c := client.NewClient(DEV)
	// 1、导入私钥恢复帐户
	piv := "46M2pAp4z3mNPuTh7jS8XHSn69TC4FAnX33Avjx7wqy3W1zzZKiSoBmNTH5PEBDKu7xR2rPa9ocSyzGWYFK7VRF2"
	alice, err := types.AccountFromBase58(piv)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	//bob := "6cPnfGr9Y4bZK7ykNpxe2hkKfaPPgsy6Tu5ahyGhzQLt"
	bob := "8pZWotbBSKBy6Luxf41TxUxDvZHp49jAbuUe2yFZRorE"

	// 查询下地址的当前余额
	balances, err := c.GetBalance(context.Background(), alice.PublicKey.String())
	if err != nil {
		fmt.Printf("get balances err = %v", err)
	}
	fmt.Printf("balances = %v\n", balances)
	// 2、组装交易
	// 2.1 拿recentBlockhashResponse
	recentBlockhashResponse, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}
	fmt.Printf("recentBlockhashResponse = %v\n", recentBlockhashResponse)

	// 最小转账金额
	minimumBalanceForRentExemption, err := c.GetMinimumBalanceForRentExemption(context.Background(), 0)
	fmt.Printf("minimumBalanceForRentExemption = %v\n", minimumBalanceForRentExemption)
	// 2.2 定义Instruction然后填充它
	ins := make([]types.Instruction, 0, 2)
	// createAccount
	//createAccountInstruction := system.CreateAccount(
	//	system.CreateAccountParam{
	//		From:     alice.PublicKey,
	//		New:      common.PublicKeyFromString(alice.PublicKey.String()),
	//		Owner:    common.StakeProgramID,
	//		Lamports: nonceAccountMinimumBalance,
	//		Space:    system.NonceAccountSize,
	//	})
	//ins = append(ins, createAccountInstruction)
	// 交易分成2个大部份 Transactions 和 Instructions 指令，Instructions可以有多个
	// 追加优先费
	ComputerUnitPrice := uint64(110000) //一个计算单元的价格
	ComputerUnitLimit := uint32(200000) //计算单元的限制
	ins = append(ins, compute_budget.SetComputeUnitPrice(compute_budget.SetComputeUnitPriceParam{
		MicroLamports: ComputerUnitPrice,
	}))

	ins = append(ins, compute_budget.SetComputeUnitLimit(compute_budget.SetComputeUnitLimitParam{
		Units: ComputerUnitLimit,
	}))
	// 这里我们先做一个简单的交易 ins就是Instructions\
	ins = append(ins, system.Transfer(
		system.TransferParam{
			From:   common.PublicKeyFromString(alice.PublicKey.String()),
			To:     common.PublicKeyFromString(bob),
			Amount: minimumBalanceForRentExemption,
		}))
	message := types.NewMessage(
		types.NewMessageParam{
			FeePayer:        common.PublicKeyFromString(alice.PublicKey.String()),
			Instructions:    ins,
			RecentBlockhash: "",
		})
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{alice},
	})
	if err != nil {
		log.Fatalf("failed to new transaction, err: %v", err)
	}
	// 3、广播上链
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to SendTransaction, err: %v", err)
	}
	fmt.Println("tx hash", txhash)
}

// SendSolMore 一笔交易里多个指令
func SendSolMore() {
	// 拆解

	c := client.NewClient(DEV)
	// 1、导入私钥恢复帐户
	alice, err := types.AccountFromBase58(Piv)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	//bob := "6cPnfGr9Y4bZK7ykNpxe2hkKfaPPgsy6Tu5ahyGhzQLt"
	//bob := "8pZWotbBSKBy6Luxf41TxUxDvZHp49jAbuUe2yFZRorE"

	// 查询下地址的当前余额
	balances, err := c.GetBalance(context.Background(), alice.PublicKey.String())
	if err != nil {
		fmt.Printf("get balances err = %v", err)
	}
	fmt.Printf("balances = %v\n", balances)
	// 2、组装交易
	// 2.1 拿recentBlockhashResponse
	recentBlockhashResponse, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}
	fmt.Printf("recentBlockhashResponse = %v\n", recentBlockhashResponse)

	// 最小转账金额
	minimumBalanceForRentExemption, err := c.GetMinimumBalanceForRentExemption(context.Background(), 0)
	fmt.Printf("minimumBalanceForRentExemption = %v\n", minimumBalanceForRentExemption)
	// 2.2 定义Instruction然后填充它
	ins := make([]types.Instruction, 0, 2)
	// 交易分成2个大部份 Transactions 和 Instructions 指令，Instructions可以有多个
	// 追加优先费
	ComputerUnitPrice := uint64(110000) //一个计算单元的价格
	ComputerUnitLimit := uint32(200000) //计算单元的限制
	ins = append(ins, compute_budget.SetComputeUnitPrice(compute_budget.SetComputeUnitPriceParam{
		MicroLamports: ComputerUnitPrice,
	}))

	ins = append(ins, compute_budget.SetComputeUnitLimit(compute_budget.SetComputeUnitLimitParam{
		Units: ComputerUnitLimit,
	}))
	// 这里我们先做一个简单的交易 ins就是 createAccount
	//createAccountIns:=types.Instruction{
	//	ProgramID: common.SystemProgramID,
	//	Accounts: "",
	//}

	message := types.NewMessage(
		types.NewMessageParam{
			FeePayer:        common.PublicKeyFromString(alice.PublicKey.String()),
			Instructions:    ins,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
		})
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{alice},
	})
	if err != nil {
		log.Fatalf("failed to new transaction, err: %v", err)
	}
	// 3、广播上链
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to SendTransaction, err: %v", err)
	}
	fmt.Println("tx hash", txhash)
}

// SendToken 用于组装Token代币的的交易
func SendToken() {

	tokenAddress := "58g1sdTgzazjR4ApXxPP63JRHj4ZDXePX251An7fxEhG"
	DEV := "https://solana-devnet.g.alchemy.com/v2/wqZxT7UnY6AgrzV42CtGgGQ7ZGM-UrTq"
	c := client.NewClient(DEV)
	// 1、获取代币的owner 拥有者
	//accInfo, err := c.GetAccountInfo(context.Background(), tokenAddress)
	//if err != nil {
	//	log.Fatalf("failed to GetAccountInfo, err: %v", err)
	//}
	//
	//fmt.Printf("accInfo Owner =%v\n", accInfo.Owner)
	// 2 寻找用户在这个token上的地址
	piv := "46M2pAp4z3mNPuTh7jS8XHSn69TC4FAnX33Avjx7wqy3W1zzZKiSoBmNTH5PEBDKu7xR2rPa9ocSyzGWYFK7VRF2"
	alice, err := types.AccountFromBase58(piv)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	bob := "6cPnfGr9Y4bZK7ykNpxe2hkKfaPPgsy6Tu5ahyGhzQLt"
	//bob := "8pZWotbBSKBy6Luxf41TxUxDvZHp49jAbuUe2yFZRorE"
	fromTokenATA, _, err := common.FindAssociatedTokenAddress(alice.PublicKey, common.PublicKeyFromString(tokenAddress))
	if err != nil {
		log.Fatalf("failed to FindAssociatedTokenAddress, err: %v", err)
	}
	//fmt.Printf("fromTokenATA=%v\n", fromTokenATA)
	toTokenATA, _, err := common.FindAssociatedTokenAddress(common.PublicKeyFromString(bob), common.PublicKeyFromString(tokenAddress))
	if err != nil {
		log.Fatalf("failed to FindAssociatedTokenAddress, err: %v", err)
	}
	//fmt.Printf("toTokenATA=%v\n", toTokenATA)
	ins := make([]types.Instruction, 0, 2)
	recentBlockhashResponse, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}
	info, err := c.GetTokenAccount(context.Background(), toTokenATA.ToBase58())
	if err != nil {
		if errors.Is(err, token.ErrInvalidAccountOwner) {
			fmt.Println("test")
			ins = append(ins, associated_token_account.Create(
				associated_token_account.CreateParam{
					Funder:                 common.PublicKeyFromString(alice.PublicKey.String()),
					Owner:                  common.PublicKeyFromString(bob),
					Mint:                   common.PublicKeyFromString(tokenAddress),
					AssociatedTokenAccount: toTokenATA,
					//ProgramID:              programID,
				}))

		} else {
			log.Fatalf("failed to GetTokenAccount, err: %v", err)
		}
	} else {
		if info.Owner.ToBase58() != bob {
			log.Fatalf("failed to GetTokenAccount,info!=bob err: %v", err)
		}
	}

	// 追加优先费
	ComputerUnitPrice := uint64(110000) //一个计算单元的价格
	ComputerUnitLimit := uint32(200000) //计算单元的限制
	ins = append(ins, compute_budget.SetComputeUnitPrice(compute_budget.SetComputeUnitPriceParam{
		MicroLamports: ComputerUnitPrice,
	}))

	ins = append(ins, compute_budget.SetComputeUnitLimit(compute_budget.SetComputeUnitLimitParam{
		Units: ComputerUnitLimit,
	}))
	ins = append(ins, token.TransferChecked(
		token.TransferCheckedParam{
			From:     fromTokenATA,
			To:       toTokenATA,
			Mint:     common.PublicKeyFromString(tokenAddress),
			Auth:     alice.PublicKey,
			Signers:  []common.PublicKey{alice.PublicKey},
			Amount:   40000000000,
			Decimals: 9, //精度

		}),
	)

	message := types.NewMessage(
		types.NewMessageParam{
			FeePayer:        common.PublicKeyFromString(alice.PublicKey.String()),
			Instructions:    ins,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
		})
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{alice},
	})

	if err != nil {
		log.Fatalf("failed to new transaction, err: %v", err)
	}
	// 3、广播上链
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to SendTransaction, err: %v", err)
	}
	fmt.Println("tx hash", txhash)
}

func RaydiumSwap() {
	// 生成客户端
	c := NewClient()
	// 导入sol用户
	piv := "46M2pAp4z3mNPuTh7jS8XHSn69TC4FAnX33Avjx7wqy3W1zzZKiSoBmNTH5PEBDKu7xR2rPa9ocSyzGWYFK7VRF2"
	alice, err := types.AccountFromBase58(piv)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	ins := make([]types.Instruction, 0, 2)
	recentBlockhashResponse, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}
	// 追加优先费
	ComputerUnitPrice := uint64(110000) //一个计算单元的价格
	ComputerUnitLimit := uint32(200000) //计算单元的限制
	ins = append(ins, compute_budget.SetComputeUnitPrice(compute_budget.SetComputeUnitPriceParam{
		MicroLamports: ComputerUnitPrice,
	}))

	ins = append(ins, compute_budget.SetComputeUnitLimit(compute_budget.SetComputeUnitLimitParam{
		Units: ComputerUnitLimit,
	}))
	// 组装swap交易的ins
	ins = append(ins)
	programID := common.PublicKeyFromString("devi51mZmdwUJGU9hjN27vEz64Gps7uUefqxg27EAtH")
	data, err := bincode.SerializeData(struct {
		Instruction Instruction
		MaxAmountIn uint64
		AmountOut   uint64
	}{
		Instruction: InstructionSwapBaseOut,
		MaxAmountIn: 1000,  // MaxAmountIn, 最大进来金额
		AmountOut:   20000, // AmountOut,输出金额
	})
	accounts := make([]types.AccountMeta, 0, 3)
	// swapV2 的account
	ammConfigAddr := "7J3nJygKrwVWyja1Q5i7aCgdjPdx2EaAJKRgEyQenhjA"
	accounts = append(accounts, types.AccountMeta{PubKey: alice.PublicKey, IsWritable: false, IsSigner: true})                            //payer
	accounts = append(accounts, types.AccountMeta{PubKey: common.PublicKeyFromString(ammConfigAddr), IsWritable: false, IsSigner: false}) //ammConfig
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //poolState
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //inputTokenAccount
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //outputTokenAccount
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //inputVault
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //outputVault
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //observationState
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //tokenProgram
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //tokenProgram2022
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //memoProgram
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //inputVaultMint
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //outputVaultMint
	swapIns := types.Instruction{
		ProgramID: programID,
		Data:      data,
		Accounts:  accounts,
	}
	ins = append(ins, swapIns)

	message := types.NewMessage(
		types.NewMessageParam{
			FeePayer:        common.PublicKeyFromString(alice.PublicKey.String()),
			Instructions:    ins,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
		})
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{alice},
	})
	if err != nil {
		log.Fatalf("failed to new transaction, err: %v", err)
	}
	// 3、广播上链
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to SendTransaction, err: %v", err)
	}
	fmt.Println("tx hash", txhash)
}

// Swap 用于sol与代币之间的swap交易
func Swap() {

}
func NewClient() *client.Client {
	//ENDPOINT := "https://rough-frequent-tent.SOLANA_MAINNET.quiknode.pro/1f710f95479f793bf564bba4ababea2c969137f6/"
	DEV := "https://solana-devnet.g.alchemy.com/v2/wqZxT7UnY6AgrzV42CtGgGQ7ZGM-UrTq"

	c := client.NewClient(DEV)
	resp, err := c.GetVersion(context.TODO())
	if err != nil {
		log.Fatalf("failed to version info, err: %v", err)
	}

	fmt.Println("version", resp.SolanaCore)
	return c
}

func GetAccount() {
	DEV := "https://solana-devnet.g.alchemy.com/v2/wqZxT7UnY6AgrzV42CtGgGQ7ZGM-UrTq"
	//c := client.NewClient(rpc.DevnetRPCEndpoint)
	c := client.NewClient(DEV)
	accountInfo, err := c.GetAccountInfo(context.Background(), "6cPnfGr9Y4bZK7ykNpxe2hkKfaPPgsy6Tu5ahyGhzQLt")
	if err != nil {
		log.Fatalf("failed to version info, err: %v", err)
	}
	fmt.Printf("accountInfo=%v\n", accountInfo)
	fmt.Printf("accountInfo Owner =%v\n", accountInfo.Owner)
	fmt.Printf("accountInfo Executable =%v\n", accountInfo.Executable)

}
func CreateAmmPool() {
	// 生成客户端
	c := NewClient()
	// 导入sol用户
	piv := "46M2pAp4z3mNPuTh7jS8XHSn69TC4FAnX33Avjx7wqy3W1zzZKiSoBmNTH5PEBDKu7xR2rPa9ocSyzGWYFK7VRF2"
	alice, err := types.AccountFromBase58(piv)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	ins := make([]types.Instruction, 0, 2)
	recentBlockhashResponse, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}
	// 追加优先费
	ComputerUnitPrice := uint64(110000) //一个计算单元的价格
	ComputerUnitLimit := uint32(200000) //计算单元的限制
	ins = append(ins, compute_budget.SetComputeUnitPrice(compute_budget.SetComputeUnitPriceParam{
		MicroLamports: ComputerUnitPrice,
	}))

	ins = append(ins, compute_budget.SetComputeUnitLimit(compute_budget.SetComputeUnitLimitParam{
		Units: ComputerUnitLimit,
	}))
	// 组装swap交易的ins
	ins = append(ins)
	programID := common.PublicKeyFromString("devi51mZmdwUJGU9hjN27vEz64Gps7uUefqxg27EAtH")
	data, err := bincode.SerializeData(struct {
		Instruction Instruction
		MaxAmountIn uint64
		AmountOut   uint64
	}{
		Instruction: InstructionSwapBaseOut,
		MaxAmountIn: 1000,  // MaxAmountIn, 最大进来金额
		AmountOut:   20000, // AmountOut,输出金额
	})
	accounts := make([]types.AccountMeta, 0, 3)
	// swapV2 的account
	ammConfigAddr := "7J3nJygKrwVWyja1Q5i7aCgdjPdx2EaAJKRgEyQenhjA"
	accounts = append(accounts, types.AccountMeta{PubKey: alice.PublicKey, IsWritable: false, IsSigner: true})                            //payer
	accounts = append(accounts, types.AccountMeta{PubKey: common.PublicKeyFromString(ammConfigAddr), IsWritable: false, IsSigner: false}) //ammConfig
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //poolState
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //inputTokenAccount
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //outputTokenAccount
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //inputVault
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //outputVault
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //observationState
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //tokenProgram
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //tokenProgram2022
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //memoProgram
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //inputVaultMint
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //outputVaultMint
	swapIns := types.Instruction{
		ProgramID: programID,
		Data:      data,
		Accounts:  accounts,
	}
	ins = append(ins, swapIns)

	message := types.NewMessage(
		types.NewMessageParam{
			FeePayer:        common.PublicKeyFromString(alice.PublicKey.String()),
			Instructions:    ins,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
		})
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{alice},
	})
	if err != nil {
		log.Fatalf("failed to new transaction, err: %v", err)
	}
	// 3、广播上链
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to SendTransaction, err: %v", err)
	}
	fmt.Println("tx hash", txhash)
}

// CreateMarket s
func CreateMarket() {
	// 生成客户端
	c := NewClient()

	// 导入sol用户
	piv := "46M2pAp4z3mNPuTh7jS8XHSn69TC4FAnX33Avjx7wqy3W1zzZKiSoBmNTH5PEBDKu7xR2rPa9ocSyzGWYFK7VRF2"
	alice, err := types.AccountFromBase58(piv)
	if err != nil {
		fmt.Printf("err=%v", err)
	}
	ins := make([]types.Instruction, 0, 2)
	recentBlockhashResponse, err := c.GetLatestBlockhash(context.Background())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}
	// 追加优先费
	ComputerUnitPrice := uint64(110000) //一个计算单元的价格
	ComputerUnitLimit := uint32(200000) //计算单元的限制
	ins = append(ins, compute_budget.SetComputeUnitPrice(compute_budget.SetComputeUnitPriceParam{
		MicroLamports: ComputerUnitPrice,
	}))

	ins = append(ins, compute_budget.SetComputeUnitLimit(compute_budget.SetComputeUnitLimitParam{
		Units: ComputerUnitLimit,
	}))
	// 组装swap交易的ins
	ins = append(ins)
	programID := common.PublicKeyFromString("devi51mZmdwUJGU9hjN27vEz64Gps7uUefqxg27EAtH")
	data, err := bincode.SerializeData(struct {
		Instruction Instruction
		MaxAmountIn uint64
		AmountOut   uint64
	}{
		Instruction: InstructionSwapBaseOut,
		MaxAmountIn: 1000,  // MaxAmountIn, 最大进来金额
		AmountOut:   20000, // AmountOut,输出金额
	})
	accounts := make([]types.AccountMeta, 0, 3)
	// swapV2 的account
	ammConfigAddr := "7J3nJygKrwVWyja1Q5i7aCgdjPdx2EaAJKRgEyQenhjA"
	accounts = append(accounts, types.AccountMeta{PubKey: alice.PublicKey, IsWritable: false, IsSigner: true})                            //payer
	accounts = append(accounts, types.AccountMeta{PubKey: common.PublicKeyFromString(ammConfigAddr), IsWritable: false, IsSigner: false}) //ammConfig
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //poolState
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //inputTokenAccount
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //outputTokenAccount
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //inputVault
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //outputVault
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: true, IsSigner: false})                      //observationState
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //tokenProgram
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //tokenProgram2022
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //memoProgram
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //inputVaultMint
	//accounts = append(accounts, types.AccountMeta{PubKey: common.TokenProgramID, IsWritable: false, IsSigner: false})                     //outputVaultMint
	swapIns := types.Instruction{
		ProgramID: programID,
		Data:      data,
		Accounts:  accounts,
	}
	ins = append(ins, swapIns)

	message := types.NewMessage(
		types.NewMessageParam{
			FeePayer:        common.PublicKeyFromString(alice.PublicKey.String()),
			Instructions:    ins,
			RecentBlockhash: recentBlockhashResponse.Blockhash,
		})
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Message: message,
		Signers: []types.Account{alice},
	})
	if err != nil {
		log.Fatalf("failed to new transaction, err: %v", err)
	}
	// 3、广播上链
	txhash, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("failed to SendTransaction, err: %v", err)
	}
	fmt.Println("tx hash", txhash)
}
