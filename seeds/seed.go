package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()

	user := getEnv("MYSQL_USER", "hobbyuser")
	pwd := getEnv("MYSQL_PWD", "hobbypass")
	host := getEnv("MYSQL_HOST", "127.0.0.1:3306")
	dbName := getEnv("MYSQL_DATABASE", "hobby_relay")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4", user, pwd, host, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}

	log.Println("Connected to database")

	if err := seedCategories(db); err != nil {
		log.Fatalf("seedCategories failed: %v", err)
	}
	if err := seedHobbies(db); err != nil {
		log.Fatalf("seedHobbies failed: %v", err)
	}
	sellerID, err := seedDemoUser(db)
	if err != nil {
		log.Fatalf("seedDemoUser failed: %v", err)
	}
	if err := seedStarterSets(db, sellerID); err != nil {
		log.Fatalf("seedStarterSets failed: %v", err)
	}

	log.Println("Seed completed successfully!")
}

func seedCategories(db *sql.DB) error {
	categories := []struct {
		name        string
		slug        string
		description string
		iconName    string
		sortOrder   int
	}{
		{"音楽", "music", "楽器演奏、作曲など音楽全般", "music_note", 1},
		{"アウトドア", "outdoor", "キャンプ、釣り、登山などのアウトドア活動", "outdoor", 2},
		{"アート・クラフト", "art-craft", "絵画、陶芸、手芸などの創作活動", "art", 3},
		{"フード・ドリンク", "food-drink", "コーヒー、ワイン、パン作りなど食に関する趣味", "coffee", 4},
		{"フィットネス・スポーツ", "fitness-sport", "ヨガ、格闘技、スポーツ全般", "fitness", 5},
		{"テクノロジー", "technology", "電子工作、プログラミング、3Dプリントなど", "computer", 6},
	}

	for _, c := range categories {
		_, err := db.Exec(
			`INSERT IGNORE INTO hobby_categories (name, slug, description, icon_name, sort_order, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			c.name, c.slug, c.description, c.iconName, c.sortOrder, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("insert category %s: %w", c.name, err)
		}
	}
	log.Printf("Seeded %d categories", len(categories))
	return nil
}

func seedHobbies(db *sql.DB) error {
	// Get category IDs
	catIDs := make(map[string]int64)
	rows, err := db.Query(`SELECT id, slug FROM hobby_categories`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var slug string
		if err := rows.Scan(&id, &slug); err != nil {
			return err
		}
		catIDs[slug] = id
	}

	hobbies := []struct {
		catSlug     string
		name        string
		slug        string
		description string
		sortOrder   int
	}{
		{"music", "ギター", "guitar", "アコースティック・エレキギターの演奏", 1},
		{"music", "ウクレレ", "ukulele", "ハワイアンな響きのウクレレ演奏", 2},
		{"music", "ドラム", "drum", "リズムを刻む打楽器演奏", 3},
		{"outdoor", "キャンプ", "camping", "大自然の中でテントを張り、焚き火を楽しむ", 1},
		{"outdoor", "釣り", "fishing", "川・海・湖での釣り全般", 2},
		{"outdoor", "登山", "hiking", "山登り・ハイキング", 3},
		{"art-craft", "イラスト", "illustration", "デジタル・アナログのイラスト制作", 1},
		{"art-craft", "水彩画", "watercolor", "水彩絵具を使った絵画", 2},
		{"food-drink", "コーヒー", "coffee", "コーヒーの自家焙煎・ハンドドリップ", 1},
		{"food-drink", "ワイン", "wine", "ワインの知識・テイスティング", 2},
		{"fitness-sport", "ヨガ", "yoga", "ヨガ・ピラティスによる心身のリラックス", 1},
		{"fitness-sport", "クライミング", "climbing", "ボルダリング・ロッククライミング", 2},
		{"technology", "電子工作", "electronics", "Arduino・Raspberry Piなどの電子工作", 1},
	}

	count := 0
	for _, h := range hobbies {
		catID, ok := catIDs[h.catSlug]
		if !ok {
			log.Printf("Category not found for slug %s, skipping hobby %s", h.catSlug, h.name)
			continue
		}
		_, err := db.Exec(
			`INSERT IGNORE INTO hobbies (category_id, name, slug, description, sort_order, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			catID, h.name, h.slug, h.description, h.sortOrder, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("insert hobby %s: %w", h.name, err)
		}
		count++
	}
	log.Printf("Seeded %d hobbies", count)
	return nil
}

func seedDemoUser(db *sql.DB) (int64, error) {
	// Check if demo user already exists
	var existing int64
	err := db.QueryRow(`SELECT id FROM users WHERE email = ?`, "demo@example.com").Scan(&existing)
	if err == nil {
		log.Printf("Demo user already exists with ID %d", existing)
		return existing, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	result, err := db.Exec(
		`INSERT INTO users (display_name, email, password_hash, avatar_url, rating_average, rating_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"デモ出品者",
		"demo@example.com",
		string(hash),
		"https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100",
		4.8,
		12,
		now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert demo user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("Created demo user with ID %d", id)
	return id, nil
}

func seedStarterSets(db *sql.DB, sellerID int64) error {
	// Get hobby IDs
	hobbyIDs := make(map[string]int64)
	catIDs := make(map[string]int64)

	rows, err := db.Query(`SELECT id, slug FROM hobbies`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var slug string
		if err := rows.Scan(&id, &slug); err != nil {
			return err
		}
		hobbyIDs[slug] = id
	}

	catRows, err := db.Query(`SELECT id, slug FROM hobby_categories`)
	if err != nil {
		return err
	}
	defer catRows.Close()
	for catRows.Next() {
		var id int64
		var slug string
		if err := catRows.Scan(&id, &slug); err != nil {
			return err
		}
		catIDs[slug] = id
	}

	now := time.Now()
	publishedAt := now

	sets := []struct {
		hobbySlug        string
		catSlug          string
		title            string
		description      string
		price            int
		beginnerScore    int
		readinessScore   int
		valueScore       int
		estimatedNew     int
		ownerNote        string
		summary          string
		imageURL         string
		items            []itemData
		recommended      []recData
	}{
		{
			hobbySlug:      "guitar",
			catSlug:        "music",
			title:          "アコースティックギター 初心者完全セット",
			description:    "ギターを始めるのに必要なものがすべて揃ったセットです。1年ほど使用しましたが、大切に保管していたため状態は良好です。チューナーや教則本も含まれており、これだけで演奏を始められます。",
			price:          12000,
			beginnerScore:  5,
			readinessScore: 95,
			valueScore:     85,
			estimatedNew:   35000,
			ownerNote:      "独学で1年練習しました。コードをある程度押さえられるようになりましたが、引越しのため手放します。",
			summary:        "このセットだけで明日からギターを弾き始められます。チューナーで正しい音に合わせ、付属のコード譜を見ながらすぐに練習できます。",
			imageURL:       "https://images.unsplash.com/photo-1510915361894-db8b60106cb1?w=600",
			items: []itemData{
				{name: "アコースティックギター（ヤマハ F310）", condition: "good", qty: 1, essential: true, note: "フレットの減りなし"},
				{name: "チューナークリップ式", condition: "like_new", qty: 1, essential: true, note: ""},
				{name: "ピック（薄・中・厚 各2枚）", condition: "new", qty: 6, essential: true, note: ""},
				{name: "カポタスト", condition: "like_new", qty: 1, essential: false, note: ""},
				{name: "ギタースタンド", condition: "good", qty: 1, essential: false, note: ""},
				{name: "コード譜入門書", condition: "good", qty: 1, essential: false, note: "書き込みなし"},
			},
			recommended: []recData{
				{name: "替え弦（ライトゲージ）", importance: "required", reason: "弦は消耗品です。切れた時に備えて1セット持っておきましょう"},
				{name: "指サック（フィンガーガード）", importance: "recommended", reason: "最初は指先が痛くなります。慣れるまであると助かります"},
				{name: "メトロノームアプリ", importance: "nice_to_have", reason: "リズム感を養うのに役立ちます。無料アプリで十分です"},
			},
		},
		{
			hobbySlug:      "coffee",
			catSlug:        "food-drink",
			title:          "本格コーヒー ハンドドリップ入門セット",
			description:    "コーヒー好きが揃えた本格的なドリップセットです。Harioのドリッパーとポットを中心に、スケールやミルも含まれています。これ一式で自宅で本格的なドリップコーヒーを楽しめます。",
			price:          8500,
			beginnerScore:  4,
			readinessScore: 90,
			valueScore:     80,
			estimatedNew:   25000,
			ownerNote:      "2年ほど使用しました。毎朝愛用していたセットです。引越しを機に新しいセットに買い替えたので出品します。",
			summary:        "このセットがあれば自宅でカフェクオリティのコーヒーを楽しめます。豆さえ用意すれば当日から本格ドリップを体験できます。",
			imageURL:       "https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?w=600",
			items: []itemData{
				{name: "Hario V60 ドリッパー（1〜2杯用）", condition: "like_new", qty: 1, essential: true, note: ""},
				{name: "Hario 細口ドリップポット 1.2L", condition: "good", qty: 1, essential: true, note: ""},
				{name: "ポーレックス コーヒーミル ミニ", condition: "good", qty: 1, essential: true, note: "刃の状態良好"},
				{name: "Hario コーヒーサーバー 600ml", condition: "like_new", qty: 1, essential: false, note: ""},
				{name: "ペーパーフィルター（100枚入）", condition: "new", qty: 1, essential: true, note: "未開封"},
				{name: "デジタルスケール（0.1g単位）", condition: "like_new", qty: 1, essential: false, note: ""},
			},
			recommended: []recData{
				{name: "コーヒー豆（200g〜）", importance: "required", reason: "まずシングルオリジンの豆を試してみてください"},
				{name: "温度計", importance: "recommended", reason: "お湯の温度で味が大きく変わります。90℃前後が目安"},
				{name: "コーヒーノート", importance: "nice_to_have", reason: "豆・挽き具合・温度を記録すると上達が早くなります"},
			},
		},
		{
			hobbySlug:      "camping",
			catSlug:        "outdoor",
			title:          "ソロキャンプ 道具一式 初心者向けセット",
			description:    "ソロキャンプを始めるために揃えた道具一式です。テントからバーナーまで、ソロキャンプに必要なものをすべてまとめました。コンパクトに収納できるので持ち運びも楽です。",
			price:          25000,
			beginnerScore:  4,
			readinessScore: 88,
			valueScore:     75,
			estimatedNew:   80000,
			ownerNote:      "10回ほど使用しました。基本的な使い方はすべて習得済みです。ファミリーキャンプに移行するため手放します。",
			summary:        "テント・シュラフ・バーナーが揃っているので、キャンプ場の予約さえすればすぐにソロキャンプを楽しめます。",
			imageURL:       "https://images.unsplash.com/photo-1504280390367-361c6d9f38f4?w=600",
			items: []itemData{
				{name: "テント（コールマン ツーリングドームST）", condition: "good", qty: 1, essential: true, note: ""},
				{name: "シュラフ（3シーズン用 -5℃対応）", condition: "good", qty: 1, essential: true, note: "洗濯済み"},
				{name: "インフレーターマット", condition: "good", qty: 1, essential: true, note: ""},
				{name: "SOTOマイクロレギュレーターストーブ", condition: "like_new", qty: 1, essential: true, note: ""},
				{name: "クッカーセット（鍋・フライパン）", condition: "good", qty: 1, essential: true, note: ""},
				{name: "ヘッドランプ（400lm）", condition: "like_new", qty: 1, essential: false, note: ""},
				{name: "折りたたみチェア", condition: "good", qty: 1, essential: false, note: ""},
			},
			recommended: []recData{
				{name: "ガス缶（OD缶 230g）", importance: "required", reason: "バーナー用燃料です。キャンプ場周辺のアウトドアショップで購入できます"},
				{name: "焚き火台", importance: "recommended", reason: "直火禁止のキャンプ場が多いため、焚き火台があると楽しみが広がります"},
				{name: "ランタン", importance: "recommended", reason: "ヘッドランプとは別に、テーブル上を照らすランタンがあると便利です"},
			},
		},
		{
			hobbySlug:      "illustration",
			catSlug:        "art-craft",
			title:          "イラスト入門セット（コピック＋スケッチブック）",
			description:    "趣味でイラストを始めた際に揃えたセットです。コピックマーカーを中心に、スケッチブックや線画用ペンも含まれています。アニメ・マンガ風イラストを描きたい方に最適なセットです。",
			price:          6000,
			beginnerScore:  5,
			readinessScore: 92,
			valueScore:     70,
			estimatedNew:   20000,
			ownerNote:      "デジタルに移行したため、アナログ道具を手放します。コピックはすべてインクが十分残っています。",
			summary:        "このセットがあれば今日からアナログイラストを始められます。コピックの基本的な使い方さえ覚えれば、すぐにカラーイラストが描けます。",
			imageURL:       "https://images.unsplash.com/photo-1513364776144-60967b0f800f?w=600",
			items: []itemData{
				{name: "コピックスケッチ 24色セット", condition: "good", qty: 1, essential: true, note: "インク残量8割以上"},
				{name: "コピック対応スケッチブック A4（30枚）", condition: "new", qty: 2, essential: true, note: "未使用"},
				{name: "ミリペン（0.1 / 0.3 / 0.5mm）", condition: "like_new", qty: 3, essential: true, note: ""},
				{name: "鉛筆セット（HB / 2B / 4B）", condition: "new", qty: 3, essential: false, note: ""},
				{name: "練り消しゴム", condition: "new", qty: 1, essential: false, note: ""},
			},
			recommended: []recData{
				{name: "コピック補充インク", importance: "required", reason: "よく使う色は消耗が早いので補充インクを用意しましょう"},
				{name: "イラスト入門書", importance: "recommended", reason: "基本的な顔・体の描き方を学ぶと上達が早まります"},
				{name: "ライトボックス", importance: "nice_to_have", reason: "下書きをなぞるときに便利です。初心者のうちはなくても大丈夫です"},
			},
		},
	}

	for _, s := range sets {
		hobbyID := hobbyIDs[s.hobbySlug]
		catID := catIDs[s.catSlug]

		// Check if already exists
		var count int
		db.QueryRow(`SELECT COUNT(*) FROM starter_sets WHERE title = ? AND seller_id = ?`, s.title, sellerID).Scan(&count)
		if count > 0 {
			log.Printf("Set '%s' already exists, skipping", s.title)
			continue
		}

		result, err := db.Exec(`
			INSERT INTO starter_sets (seller_id, hobby_id, category_id, title, description, price, status,
			                         beginner_score, readiness_score, value_score, estimated_new_price,
			                         previous_owner_note, startable_summary, published_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			sellerID, hobbyID, catID, s.title, s.description, s.price, "on_sale",
			s.beginnerScore, s.readinessScore, s.valueScore, s.estimatedNew,
			s.ownerNote, s.summary, publishedAt, now, now,
		)
		if err != nil {
			return fmt.Errorf("insert set %s: %w", s.title, err)
		}

		setID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// Insert image
		_, err = db.Exec(`INSERT INTO set_images (starter_set_id, image_url, sort_order, created_at) VALUES (?, ?, ?, ?)`,
			setID, s.imageURL, 1, now)
		if err != nil {
			return fmt.Errorf("insert image for set %d: %w", setID, err)
		}

		// Insert items
		for _, item := range s.items {
			_, err = db.Exec(`INSERT INTO set_items (starter_set_id, name, condition_label, quantity, is_essential, note, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				setID, item.name, item.condition, item.qty, item.essential, item.note, now, now)
			if err != nil {
				return fmt.Errorf("insert item %s for set %d: %w", item.name, setID, err)
			}
		}

		// Insert recommended items
		for _, rec := range s.recommended {
			_, err = db.Exec(`INSERT INTO recommended_items (starter_set_id, name, importance, reason, created_at) VALUES (?, ?, ?, ?, ?)`,
				setID, rec.name, rec.importance, rec.reason, now)
			if err != nil {
				return fmt.Errorf("insert rec item %s for set %d: %w", rec.name, setID, err)
			}
		}

		log.Printf("Seeded set '%s' (ID: %d)", s.title, setID)
	}

	return nil
}

type itemData struct {
	name      string
	condition string
	qty       int
	essential bool
	note      string
}

type recData struct {
	name       string
	importance string
	reason     string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
