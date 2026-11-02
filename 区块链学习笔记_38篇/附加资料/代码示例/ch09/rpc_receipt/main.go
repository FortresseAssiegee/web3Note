package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

type Log struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
	Removed bool     `json:"removed"`
}
type Receipt struct {
	TransactionHash   string `json:"transactionHash"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	Status            string `json:"status"`
	GasUsed           string `json:"gasUsed"`
	EffectiveGasPrice string `json:"effectiveGasPrice"`
	Logs              []Log  `json:"logs"`
}

func receipt(ctx context.Context, endpoint, tx string) (*Receipt, error) {
	request := struct {
		JSONRPC string   `json:"jsonrpc"`
		ID      int      `json:"id"`
		Method  string   `json:"method"`
		Params  []string `json:"params"`
	}{"2.0", 1, "eth_getTransactionReceipt", []string{tx}}
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > 1<<20 {
		return nil, errors.New("response too large")
	}
	var reply struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int             `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err = json.Unmarshal(body, &reply); err != nil {
		return nil, err
	}
	if reply.JSONRPC != "2.0" || reply.ID != 1 {
		return nil, errors.New("unexpected RPC envelope")
	}
	if reply.Error != nil {
		return nil, fmt.Errorf("RPC %d: %s", reply.Error.Code, reply.Error.Message)
	}
	if bytes.Equal(bytes.TrimSpace(reply.Result), []byte("null")) {
		return nil, nil
	}
	var result Receipt
	if err = json.Unmarshal(reply.Result, &result); err != nil {
		return nil, err
	}
	if !strings.EqualFold(result.TransactionHash, tx) {
		return nil, errors.New("transaction hash mismatch")
	}
	return &result, nil
}

func quantity(s string) (*big.Int, error) {
	if !strings.HasPrefix(s, "0x") || len(s) < 3 {
		return nil, errors.New("bad hex quantity")
	}
	if len(s) > 3 && s[2] == '0' {
		return nil, errors.New("non-canonical leading zero")
	}
	for _, c := range s[2:] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return nil, errors.New("bad hex digit")
		}
	}
	n, ok := new(big.Int).SetString(s[2:], 16)
	if !ok || n.Sign() < 0 {
		return nil, errors.New("bad number")
	}
	return n, nil
}

const transferTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

func erc20Amount(l Log, expectedToken string) (*big.Int, error) {
	if l.Removed || !strings.EqualFold(l.Address, expectedToken) || len(l.Topics) != 3 || !strings.EqualFold(l.Topics[0], transferTopic) {
		return nil, errors.New("wrong emitter or event shape")
	}
	for _, topic := range l.Topics[1:] {
		if !strings.HasPrefix(topic, "0x") || len(topic) != 66 {
			return nil, errors.New("bad address topic")
		}
		word, err := hex.DecodeString(topic[2:])
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(word[:12], make([]byte, 12)) {
			return nil, errors.New("non-canonical address word")
		}
	}
	if !strings.HasPrefix(l.Data, "0x") || len(l.Data) != 66 {
		return nil, errors.New("bad uint256 word")
	}
	data, err := hex.DecodeString(l.Data[2:])
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(data), nil
}

func check(name string, ok bool) {
	fmt.Printf("%s: %v\n", name, ok)
	if !ok {
		panic(name)
	}
}

func main() {
	// A local fixture makes the lesson reproducible without public RPC or API keys.
	tx := "0x" + strings.Repeat("a", 64)
	token := "0x" + strings.Repeat("1", 40)
	l := Log{Address: token, Topics: []string{transferTopic, "0x" + strings.Repeat("0", 24) + strings.Repeat("2", 40), "0x" + strings.Repeat("0", 24) + strings.Repeat("3", 40)}, Data: fmt.Sprintf("0x%064x", 1000000000000000000)}
	fixture := Receipt{TransactionHash: tx, BlockHash: "0x" + strings.Repeat("b", 64), BlockNumber: "0x10", Status: "0x1", GasUsed: "0xea60", EffectiveGasPrice: "0x3b9aca00", Logs: []Log{l}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string   `json:"method"`
			Params []string `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if req.Method != "eth_getTransactionReceipt" {
			io.WriteString(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"read method only"}}`)
			return
		}
		var result any
		if len(req.Params) == 1 && req.Params[0] == tx {
			result = fixture
		}
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": result})
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := receipt(ctx, server.URL, tx)
	if err != nil || r == nil {
		panic("fixture receipt missing")
	}
	check("receipt reports successful execution", r.Status == "0x1")
	gas, err := quantity(r.GasUsed)
	if err != nil {
		panic(err)
	}
	price, err := quantity(r.EffectiveGasPrice)
	if err != nil {
		panic(err)
	}
	fmt.Println("execution fee in wei:", new(big.Int).Mul(gas, price))
	amount, err := erc20Amount(r.Logs[0], token)
	if err != nil {
		panic(err)
	}
	fmt.Println("transfer base units:", amount)
	forged := r.Logs[0]
	forged.Address = "0x" + strings.Repeat("4", 40)
	_, err = erc20Amount(forged, token)
	check("same topic from wrong contract rejected", err != nil)
	missing, err := receipt(ctx, server.URL, "0x"+strings.Repeat("c", 64))
	check("null result kept distinct", err == nil && missing == nil)
	_, leadingZero := quantity("0x00")
	_, signed := quantity("0x+1")
	check("non-canonical quantities rejected", leadingZero != nil && signed != nil)
}
