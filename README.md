# DPoSI - Delegated Proof of Speed Internet

**DPoSI** adalah mekanisme konsensus blockchain inovatif yang menggabungkan **Delegated Proof of Stake (DPoS)** dengan **Quality of Service (QoS)** jaringan sebagai penentu utama hak produksi blok.

## Konsep Inti

### 1. Skor Kecepatan Dinamis
Setiap validator dinilai berdasarkan kualitas jaringan mereka:

```
Score = (0.4 × Throughput_Mbps) + (0.4 × (1000 / Latency_ms)) + (0.2 × Uptime_Percentage)
```

**Komponen:**
- **Throughput (40%)**: Kecepatan download & upload rata-rata
- **Latency (40%)**: Latensi jaringan (semakin rendah semakin baik)
- **Uptime (20%)**: Persentase waktu node online dalam 30 hari terakhir

### 2. Struktur Siklus
- **1 Siklus** = **100 node aktif** × **10 blok per node** = **1000 blok total**
- Setiap node mendapat giliran tetap 10 blok dalam satu siklus
- Rotasi menggunakan **Deterministic Shuffled Round-Robin**

### 3. Rotasi Cerdas
Blok-blok dari satu node **tersebar di seluruh siklus** (bukan berurutan):
- **Keuntungan**: Jika node offline, hanya 1 blok terlewat, bukan 10 berturut-turut
- **Dampak**: Meningkatkan resiliensi dan fairness

### 4. Speed Testing Real-time
- Setiap **10 blok**, micro-speed-test dijalankan untuk **5 node acak**
- Hasil verifikasi silang oleh 3 validator independen
- Skor diperbarui menggunakan **Exponential Moving Average (EMA)**

## Struktur Proyek

```
dposi-blockchain/
├── cmd/
│   └── dposi-node/              # Entrypoint utama
│       └── main.go
├── internal/
│   ├── consensus/               # Konsensus engine
│   │   ├── engine.go
│   │   ├── scheduler.go
│   │   ├── validator.go
│   │   └── slashing.go
│   ├── speed/                   # Modul speed testing
│   │   ├── tester.go
│   │   ├── scorer.go
│   │   └── verifier.go
│   ├── state/                   # State management
│   │   ├── cycle.go
│   │   ├── candidate_pool.go
│   │   └── snapshot.go
│   └── storage/                 # Database wrapper
│       ├── block_store.go
│       ├── state_db.go
│       └── db.go
├── pkg/
│   ├── types/                   # Type definitions
│   │   ├── block.go
│   │   ├── transaction.go
│   │   ├── node.go
│   │   └── config.go
│   ├── crypto/                  # Cryptography
│   │   ├── ed25519.go
│   │   └── hash.go
│   ├── p2p/                     # Networking (libp2p)
│   │   ├── peer_discovery.go
│   │   ├── broadcaster.go
│   │   └── speed_protocol.go
│   └── utils/                   # Utilities
│       ├── logger.go
│       └── math_utils.go
├── configs/
│   ├── genesis.json             # Genesis block config
│   └── config.toml              # Node config
├── scripts/
│   ├── build.sh
│   ├── test_local.sh
│   └── benchmark/
├── go.mod
├── go.sum
└── README.md
```

## Development Roadmap

### Sprint 1: Fondasi Data & Kripto ✅
- [x] Definisi Block, Transaction, NodeCandidate
- [x] Ed25519 signing & verification
- [x] BadgerDB storage layer

### Sprint 2: Modul Kecepatan & Scoring ⏳
- [ ] SpeedTester: TCP speed test ke 5 node acak
- [ ] Scorer: Hitung dynamic score
- [ ] Verifier: Cross-check hasil test

### Sprint 3: Scheduler & Konsensus ⏳
- [ ] Deterministic Shuffled Round-Robin scheduler
- [ ] Consensus engine (main loop)
- [ ] Slashing logic

### Sprint 4: P2P & Sinkronisasi ⏳
- [ ] libp2p integration
- [ ] Gossip protocol
- [ ] Block sync & finality

## Quick Start

### Prerequisites
- Go 1.21+
- BadgerDB v4
- libp2p

### Build
```bash
go mod download
go build -o dposi-node ./cmd/dposi-node
```

### Run Node
```bash
./dposi-node --config configs/config.toml
```

### Testing (Local Network)
```bash
bash scripts/test_local.sh
```

## Security Features

1. **Anti-Sybil Speed Testing**: Verifikasi silang result speed test setiap 10 blok
2. **Geolocation Fairness**: Bonus handicap untuk region dengan latensi tinggi
3. **Slashing**: Penalti otomatis untuk node offline atau double-signing
4. **Database Pruning**: Keep hanya 3 siklus terakhir untuk efisiensi storage

## Performance Metrics

- **Block Time**: 3 detik
- **Finality**: ~30 detik (10 blok confirmations)
- **Throughput**: Target 1000+ Tx/sec (disesuaikan dengan network)
- **Validators**: 100 aktif per siklus

## License

MIT License - Lihat LICENSE file untuk detail

## Contributing

Kontribusi welcome! Silakan buat issue atau pull request.
