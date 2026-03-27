package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/net/proxy" // Necessário: go get golang.org/x/net/proxy
)

const (
	CertDir      = "/etc/pixelcraft/certs"
	PortMin      = 50000
	PortMax      = 51000
	PortCount    = 10
	RotationSec  = 30
	ProxyAddr    = "127.0.0.1:40000" // Porta padrão do WARP Proxy
)

type BankTransaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	Ref    string `json:"ref"`
}

func main() {
	targetIP := flag.String("ip", "10.0.0.5", "IP do AxionPay na rede Zero Trust")
	secret := flag.String("secret", "", "HYDRA_SECRET para rotação de portas")
	flag.Parse()

	if *secret == "" {
		log.Fatal("❌ Erro: O HYDRA_SECRET é obrigatório. Use -secret <chave>")
	}

	log.Println("⚔️  Iniciando PixelCraft Hydra Client (via WARP Proxy)...")

	// 1. Configurar mTLS (Identidade Digital)
	cert, err := tls.LoadX509KeyPair(CertDir+"/client.crt", CertDir+"/client.key")
	if err != nil {
		log.Fatalf("❌ Falha nos certs: %v", err)
	}

	caCert, err := os.ReadFile(CertDir + "/ca.crt")
	if err != nil {
		log.Fatalf("❌ Falha na CA: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true, // Hostname não importa em IP privado
	}

	// 2. Criar Dialer SOCKS5 (O Pulo do Gato para não cair o SSH)
	dialer, err := proxy.SOCKS5("tcp", ProxyAddr, nil, proxy.Direct)
	if err != nil {
		log.Fatalf("❌ Erro ao criar dialer do Proxy: %v", err)
	}

	// 3. Calcular Portas Hydra
	timestamp := time.Now().Unix() / RotationSec
	ports := calculatePorts([]byte(*secret), timestamp, PortMin, PortMax, PortCount)
	log.Printf("🔓 Portas Ativas (Sincronizadas): %v", ports)

	// 4. Preparar Payload
	tx := BankTransaction{
		From:   "pixelcraft_wallet_01",
		To:     "system_treasury",
		Amount: 500000,
		Ref:    fmt.Sprintf("TX-%d", time.Now().Unix()),
	}
	txJson, _ := json.Marshal(tx)
	sessionID := fmt.Sprintf("%016x", time.Now().UnixNano())

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0

	// 5. Executar Conexões Paralelas
	for i, port := range ports {
		wg.Add(1)
		go func(idx int, p int) {
			defer wg.Done()

			// Conecta via Proxy -> TCP -> mTLS
			rawConn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", *targetIP, p))
			if err != nil {
				log.Printf("   ❌ Porta %d recusada: %v", p, err)
				return
			}

			// Envelopa com TLS
			conn := tls.Client(rawConn, tlsConfig)
			defer conn.Close()

			// Monta o Frame Hydra
			frame := append([]byte(sessionID), make([]byte, 4)...)
			data := []byte(fmt.Sprintf("FRAG-%d|%s", idx, txJson))
			binary.BigEndian.PutUint32(frame[16:20], uint32(len(data)))
			frame = append(frame, data...)

			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if _, err := conn.Write(frame); err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
				fmt.Printf("   ✅ Fragmento %d injetado na porta %d\n", idx, p)
			}
		}(i, port)
	}

	wg.Wait()

	if successCount == PortCount {
		log.Println("🏆 SUCESSO! Transação invisível concluída com 10/10 fragmentos.")
	} else {
		log.Printf("⚠️  ALERTA: Apenas %d/10 fragmentos enviados. O servidor pode rejeitar.", successCount)
	}
}

func calculatePorts(secret []byte, counter int64, min, max, count int) []int {
	mac := hmac.New(sha256.New, secret)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))
	mac.Write(buf)
	seed := mac.Sum(nil)

	portRange := max - min + 1
	selected := make(map[int]bool)
	var ports []int

	offset := 0
	for len(ports) < count {
		if offset+4 > len(seed) {
			mac.Reset()
			mac.Write(seed)
			mac.Write([]byte{byte(len(ports))})
			seed = mac.Sum(nil)
			offset = 0
		}
		val := binary.BigEndian.Uint32(seed[offset : offset+4])
		port := min + int(val%uint32(portRange))
		offset += 4
		if !selected[port] {
			selected[port] = true
			ports = append(ports, port)
		}
	}
	sort.Ints(ports)
	return ports
}