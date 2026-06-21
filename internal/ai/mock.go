package ai

import (
	"fmt"
	"strings"

	"hobby-relay-backend/internal/domain"
)

type MockClient struct{}

func NewMockClient() Client {
	return &MockClient{}
}

func (m *MockClient) StartListingSupport(hobbyText string) (*ListingSupportResult, error) {
	chips := suggestedItemsForHobby(hobbyText)
	return &ListingSupportResult{
		Message:        fmt.Sprintf("「%s」のスターターセットを出品しましょう！まず、セットに含まれるアイテムを教えてください。どんなものが入っていますか？", hobbyText),
		SuggestedChips: chips,
		Progress:       domain.ProgressDTO{Current: 1, Total: 4},
		Done:           false,
	}, nil
}

func (m *MockClient) NextListingStep(sessionMessages []SessionMessage, userMessage string) (*ListingSupportResult, error) {
	// Determine current step from message history
	step := countUserMessages(sessionMessages) + 1

	switch step {
	case 2:
		// Ask about conditions
		return &ListingSupportResult{
			Message:        "ありがとうございます！それぞれのアイテムの状態を教えてください。全体的にどのくらいの状態ですか？",
			SuggestedChips: []string{"ほぼ新品", "目立った傷なし", "やや傷あり", "動作未確認"},
			Progress:       domain.ProgressDTO{Current: 2, Total: 4},
			Done:           false,
		}, nil

	case 3:
		// Ask about price
		return &ListingSupportResult{
			Message:        "状態を教えていただきありがとうございます。ご希望の販売価格を教えてください。",
			SuggestedChips: []string{"3000円", "5000円", "8000円", "10000円", "15000円"},
			Progress:       domain.ProgressDTO{Current: 3, Total: 4},
			Done:           false,
		}, nil

	case 4:
		// Ask for previous owner note
		return &ListingSupportResult{
			Message:        "価格を設定しました。購入者へのメッセージや使用感など、一言添えていただけますか？（任意）",
			SuggestedChips: []string{"大切に使っていました", "ほぼ未使用です", "初心者でも始めやすいです", "一式揃っています"},
			Progress:       domain.ProgressDTO{Current: 4, Total: 4},
			Done:           false,
		}, nil

	default:
		// Step 5+: Generate final result
		hobbyText := extractHobbyFromMessages(sessionMessages)
		itemsText := extractFirstUserMessage(sessionMessages)
		title := generateTitle(hobbyText, itemsText)
		description := generateDescription(hobbyText, itemsText)
		items := generateItems(hobbyText)
		recommended := generateRecommended(hobbyText)

		summary := "このセットだけですぐに始められます。必要なものがすべて揃っているので、届いた日から楽しめます。"
		if hobbyText != "" {
			summary = fmt.Sprintf("このセットだけで%sをすぐに始められます。必要なものがすべて揃っているので、届いた日から楽しめます。", hobbyText)
		}

		return &ListingSupportResult{
			Message:           fmt.Sprintf("セット情報の入力が完了しました！「%s」の出品準備ができています。内容を確認して出品してください。", title),
			SuggestedChips:    []string{},
			Progress:          domain.ProgressDTO{Current: 4, Total: 4},
			Done:              true,
			Title:             title,
			Description:       description,
			HobbyText:         hobbyText,
			Items:             items,
			RecommendedItems:  recommended,
			BeginnerScore:     4,
			ReadinessScore:    85,
			PreviousOwnerNote: extractLastUserMessage(sessionMessages),
			StartableSummary:  summary,
		}, nil
	}
}

func (m *MockClient) AnswerSetQuestion(ctx SetQuestionContext, userMessage string) (string, error) {
	msg := userMessage

	// 初心者・難易度
	if contains(msg, "初心者", "難しい", "大丈夫", "できる", "始められる") {
		note := ""
		if ctx.StartableSummary != "" {
			note = " " + ctx.StartableSummary
		} else if ctx.ReadinessScore >= 80 {
			note = " 届いたその日から始められる内容が揃っています。"
		}
		return fmt.Sprintf("「%s」は初心者の方でも大丈夫なセットです。%s", ctx.Title, note), nil
	}

	// 価格・新品との比較
	if contains(msg, "安い", "新品", "価格", "値段", "お得") {
		if ctx.EstimatedNewPrice > 0 {
			diff := ctx.EstimatedNewPrice - ctx.Price
			pct := diff * 100 / ctx.EstimatedNewPrice
			return fmt.Sprintf("新品で揃えると約¥%s相当のところ、このセットは¥%sです。約%d%%お得（¥%s節約）になります。",
				fmtPrice(ctx.EstimatedNewPrice), fmtPrice(ctx.Price), pct, fmtPrice(diff)), nil
		}
		return fmt.Sprintf("「%s」は¥%sで出品中です。新品より大幅にお得にスタートできます。", ctx.Title, fmtPrice(ctx.Price)), nil
	}

	// 状態・コンディション
	if contains(msg, "状態", "コンディション", "傷", "汚れ", "使用感") {
		if len(ctx.Items) == 0 {
			return "セット内アイテムの状態は商品詳細ページに記載されています。", nil
		}
		var parts []string
		for _, item := range ctx.Items {
			if item.Essential {
				label := conditionLabel(item.Condition)
				parts = append(parts, fmt.Sprintf("%s：%s", item.Name, label))
			}
		}
		if len(parts) == 0 {
			return "商品詳細ページで各アイテムの状態をご確認いただけます。", nil
		}
		return fmt.Sprintf("主なアイテムの状態です。%s。詳細は商品ページをご覧ください。", strings.Join(parts, "／")), nil
	}

	// 何が含まれているか
	if contains(msg, "含まれ", "付属", "セット内容", "何が入", "一覧") {
		if len(ctx.Items) > 0 {
			var names []string
			for _, item := range ctx.Items {
				names = append(names, item.Name)
			}
			return fmt.Sprintf("このセットには「%s」が含まれています。", strings.Join(names, "、")), nil
		}
		return "セット内容は商品詳細ページをご確認ください。", nil
	}

	// 追加で必要なもの
	if contains(msg, "必要", "他に", "追加", "買い足") {
		if len(ctx.RecommendedItems) > 0 {
			var names []string
			for _, r := range ctx.RecommendedItems {
				names = append(names, r.Name)
			}
			return fmt.Sprintf("追加であると安心なアイテムとして「%s」などが挙げられます。", strings.Join(names, "、")), nil
		}
		return "このセットだけで始めるのに十分な内容が揃っています。", nil
	}

	// すぐ始められるか
	if contains(msg, "すぐ", "今日から", "いつから", "どれくらい") {
		if ctx.ReadinessScore >= 90 {
			return fmt.Sprintf("「%s」はすぐに始めやすさスコアが%d/100と非常に高く、届いたその日から始められます。", ctx.Title, ctx.ReadinessScore), nil
		} else if ctx.ReadinessScore >= 75 {
			return fmt.Sprintf("「%s」のすぐに始めやすさスコアは%d/100です。少し準備が必要な部分もありますが、すぐに使い始められます。", ctx.Title, ctx.ReadinessScore), nil
		}
		return fmt.Sprintf("「%s」のすぐに始めやすさスコアは%d/100です。一部追加で準備が必要なアイテムがあります。", ctx.Title, ctx.ReadinessScore), nil
	}

	// 前の持ち主のコメント
	if contains(msg, "使用感", "使ってた", "前の人", "出品者", "どんな人") {
		if ctx.PreviousOwnerNote != "" {
			return fmt.Sprintf("前の持ち主より：「%s」", ctx.PreviousOwnerNote), nil
		}
	}

	// 送料
	if contains(msg, "送料") {
		return "送料は取引成立後に出品者とご相談ください。", nil
	}

	// デフォルト：セットの説明を使う
	if ctx.Description != "" {
		// 説明文の先頭80文字程度を使う
		desc := ctx.Description
		if len([]rune(desc)) > 80 {
			runes := []rune(desc)
			desc = string(runes[:80]) + "…"
		}
		return fmt.Sprintf("%s ご不明な点があればお気軽にどうぞ。", desc), nil
	}
	return fmt.Sprintf("「%s」についてのご質問ありがとうございます。詳細は商品ページをご確認いただくか、出品者にメッセージでお問い合わせください。", ctx.Title), nil
}

func contains(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func conditionLabel(c string) string {
	switch c {
	case "new":
		return "新品"
	case "like_new":
		return "ほぼ新品"
	case "good":
		return "良好"
	case "fair":
		return "やや使用感あり"
	default:
		return "状態不明"
	}
}

func fmtPrice(p int) string {
	s := fmt.Sprintf("%d", p)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

func (m *MockClient) InterpretSearchQuery(query string) (*SearchInterpretation, error) {
	result := &SearchInterpretation{}

	// Price detection
	if strings.Contains(query, "1万円以内") || strings.Contains(query, "10000円以内") || strings.Contains(query, "１万円以内") {
		result.MaxPrice = 10000
	} else if strings.Contains(query, "5000円以内") || strings.Contains(query, "５０００円以内") {
		result.MaxPrice = 5000
	} else if strings.Contains(query, "3000円以内") || strings.Contains(query, "３０００円以内") {
		result.MaxPrice = 3000
	} else if strings.Contains(query, "15000円以内") {
		result.MaxPrice = 15000
	} else if strings.Contains(query, "2万円以内") || strings.Contains(query, "20000円以内") {
		result.MaxPrice = 20000
	}

	// Beginner detection
	if strings.Contains(query, "初心者") || strings.Contains(query, "はじめて") || strings.Contains(query, "入門") {
		result.MinBeginnerScore = 4
	}

	// Readiness detection
	if strings.Contains(query, "すぐ始められる") || strings.Contains(query, "すぐに始め") {
		result.MinReadinessScore = 80
	}

	// Hobby/category detection
	if strings.Contains(query, "家でできる") || strings.Contains(query, "室内") || strings.Contains(query, "インドア") {
		result.RelatedHobbies = []string{"コーヒー", "イラスト", "ヨガ", "ウクレレ"}
		result.SmartMessage = "「家でできる」趣味のセットを表示しています。"
	} else if strings.Contains(query, "アウトドア") || strings.Contains(query, "外") {
		result.RelatedHobbies = []string{"キャンプ", "釣り"}
		result.SmartMessage = "アウトドア系のセットを表示しています。"
	} else if strings.Contains(query, "音楽") {
		result.RelatedHobbies = []string{"ギター", "ウクレレ", "ピアノ", "ドラム"}
		result.SmartMessage = "音楽系のセットを表示しています。"
	} else if strings.Contains(query, "アート") || strings.Contains(query, "絵") {
		result.RelatedHobbies = []string{"イラスト", "水彩画", "油絵"}
		result.SmartMessage = "アート・絵画系のセットを表示しています。"
	} else if strings.Contains(query, "料理") || strings.Contains(query, "食") {
		result.RelatedHobbies = []string{"コーヒー", "パン作り", "ワイン"}
		result.SmartMessage = "料理・食文化系のセットを表示しています。"
	} else if result.SmartMessage == "" {
		result.SmartMessage = fmt.Sprintf("「%s」に近いセットを表示しています。", query)
	}

	return result, nil
}

func (m *MockClient) GenerateStartPlan(setTitle string, hobbyName string) ([]domain.StartPlanStepDTO, error) {
	plans := map[string][]domain.StartPlanStepDTO{
		"ギター": {
			{DayNo: 1, Title: "ギターに触れてみよう", Body: "まずはギターを手に取り、各部の名称を覚えましょう。チューナーを使って正しいチューニングをしてみてください。"},
			{DayNo: 2, Title: "基本コードを覚えよう（Am・Em）", Body: "マイナーコード2つから始めます。AmとEmは指の負担が少なく、押さえやすいコードです。"},
			{DayNo: 3, Title: "基本コードを覚えよう（C・G）", Body: "よく使うCとGコードを練習します。左手の指先が慣れてきたら音が出やすくなります。"},
			{DayNo: 4, Title: "ストロークパターンを練習しよう", Body: "ダウンストローク→アップストロークの基本リズムを練習します。メトロノームを使いましょう。"},
			{DayNo: 5, Title: "コードチェンジに挑戦", Body: "Am→Em→C→Gの順番でコードチェンジを練習。最初はゆっくりで大丈夫です。"},
			{DayNo: 6, Title: "簡単な曲にチャレンジ", Body: "「ハッピーバースデー」など簡単な曲のコード譜を探して弾いてみましょう。"},
			{DayNo: 7, Title: "振り返りと次のステップ", Body: "1週間の練習を振り返りましょう。毎日少しずつ練習することが上達への近道です。"},
		},
		"コーヒー": {
			{DayNo: 1, Title: "コーヒー器具を整えよう", Body: "セットに含まれる器具を確認し、それぞれの役割を理解しましょう。まずは清潔に保つことが大切です。"},
			{DayNo: 2, Title: "豆の種類と挽き方を学ぼう", Body: "コーヒー豆の産地と焙煎度の違いを知りましょう。グラインダーの粗さ調整を試してみてください。"},
			{DayNo: 3, Title: "初めてのドリップコーヒー", Body: "ペーパードリップで一杯淹れてみましょう。お湯の温度は90℃前後が目安です。"},
			{DayNo: 4, Title: "蒸らしのコツをつかもう", Body: "注いだ後30秒蒸らすと、香りと味が引き立ちます。蒸らしありとなしを比較してみてください。"},
			{DayNo: 5, Title: "お湯の量とスピードを調整しよう", Body: "細口ポットを使って、お湯を細くゆっくり注ぐ練習をしましょう。円を描くように注ぎます。"},
			{DayNo: 6, Title: "豆を変えて飲み比べ", Body: "異なる産地の豆で淹れ比べをしてみましょう。エチオピアとブラジルでは味の違いを感じられます。"},
			{DayNo: 7, Title: "自分好みの一杯を見つけよう", Body: "これまでの記録を見直し、自分好みのレシピを確立しましょう。コーヒーノートをつけるとさらに上達します。"},
		},
		"釣り": {
			{DayNo: 1, Title: "釣り具の確認とセッティング", Body: "ロッド、リール、ラインの組み合わせを確認しましょう。リールへのライン巻き取りを練習します。"},
			{DayNo: 2, Title: "基本的な結び方をマスター", Body: "ユニノットやパロマーノットなど基本の結び方を練習します。家で何度も繰り返しましょう。"},
			{DayNo: 3, Title: "キャスティング練習（陸上）", Body: "公園など広い場所でキャスティングを練習します。狙った場所に投げられるよう繰り返しましょう。"},
			{DayNo: 4, Title: "釣り場リサーチ", Body: "近くの釣りスポットを調べましょう。初心者には管理釣り場がおすすめです。必要な遊漁券も確認を。"},
			{DayNo: 5, Title: "いざ実釣へ！", Body: "実際に釣り場へ行ってみましょう。まずはウキ釣りで魚のアタリを感じることを目標にしてください。"},
			{DayNo: 6, Title: "釣果を振り返ろう", Body: "釣れた魚、使ったエサや仕掛けを記録しましょう。次回への改善点を考えます。"},
			{DayNo: 7, Title: "次のステップを計画しよう", Body: "釣りの技術は経験が大切です。定期的に釣りに行く計画を立て、釣り仲間を見つけるのもいいでしょう。"},
		},
		"キャンプ": {
			{DayNo: 1, Title: "ギアを確認しよう", Body: "テント、シュラフ、マット、バーナーなどを広げて確認します。説明書を読んで各パーツを理解しましょう。"},
			{DayNo: 2, Title: "テントの設営練習", Body: "自宅の庭やリビングでテントを張る練習をします。30分以内に設営できるようになりましょう。"},
			{DayNo: 3, Title: "バーナーでお湯を沸かそう", Body: "ガスバーナーの使い方を練習します。お湯を沸かしてカップラーメンを作ってみましょう。"},
			{DayNo: 4, Title: "キャンプ場リサーチ", Body: "近くのキャンプ場を調べて予約しましょう。初心者には設備の整ったオートキャンプ場がおすすめです。"},
			{DayNo: 5, Title: "パッキングを練習しよう", Body: "必要なものをリスト化し、効率よくパッキングする練習をします。重いものを下に入れるのがコツです。"},
			{DayNo: 6, Title: "いざキャンプへ！", Body: "いよいよ初キャンプです。テント設営から就寝まで、楽しみながら学びましょう。"},
			{DayNo: 7, Title: "振り返りと次回計画", Body: "初回キャンプの感想をまとめましょう。次回に向けて追加で必要なものをリストアップします。"},
		},
	}

	if steps, ok := plans[hobbyName]; ok {
		return steps, nil
	}

	// Default generic plan
	return []domain.StartPlanStepDTO{
		{DayNo: 1, Title: fmt.Sprintf("%sを始めよう", hobbyName), Body: fmt.Sprintf("「%s」のセットを開封して、含まれているアイテムをすべて確認しましょう。", setTitle)},
		{DayNo: 2, Title: "基礎知識を学ぼう", Body: fmt.Sprintf("%sに関する入門書や動画を参考に、基本的な知識を身につけましょう。", hobbyName)},
		{DayNo: 3, Title: "基本動作を練習しよう", Body: "まずは基本の動作や操作をゆっくりと練習します。完璧でなくても大丈夫です。"},
		{DayNo: 4, Title: "実践してみよう", Body: "学んだことを実際に試してみましょう。失敗を恐れずに挑戦することが大切です。"},
		{DayNo: 5, Title: "コツをつかもう", Body: "繰り返し練習することで徐々にコツがつかめてきます。小さな成功体験を積み重ねましょう。"},
		{DayNo: 6, Title: "応用にチャレンジ", Body: "基本ができてきたら、少し難しいことに挑戦してみましょう。"},
		{DayNo: 7, Title: "振り返りと次のステップ", Body: fmt.Sprintf("1週間の%sを振り返り、次のステップを計画しましょう。継続することが上達の秘訣です。", hobbyName)},
	}, nil
}

// Helper functions

func countUserMessages(messages []SessionMessage) int {
	count := 0
	for _, m := range messages {
		if m.Sender == "user" {
			count++
		}
	}
	return count
}

func extractHobbyFromMessages(messages []SessionMessage) string {
	for _, m := range messages {
		if m.Sender == "system" {
			// Format: "趣味: hobbyText"
			if strings.HasPrefix(m.Message, "趣味:") || strings.HasPrefix(m.Message, "趣味：") {
				parts := strings.SplitN(m.Message, ":", 2)
				if len(parts) == 2 {
					hobby := strings.TrimSpace(parts[1])
					if hobby != "" {
						return hobby
					}
				}
				// full-width colon fallback
				parts2 := strings.SplitN(m.Message, "：", 2)
				if len(parts2) == 2 {
					hobby := strings.TrimSpace(parts2[1])
					if hobby != "" {
						return hobby
					}
				}
			}
			// Quoted format: 「hobbyText」
			qparts := strings.Split(m.Message, "「")
			if len(qparts) > 1 {
				qparts2 := strings.Split(qparts[1], "」")
				if len(qparts2) > 0 && qparts2[0] != "" {
					return qparts2[0]
				}
			}
		}
	}
	return ""
}

func extractFirstUserMessage(messages []SessionMessage) string {
	for _, m := range messages {
		if m.Sender == "user" {
			return m.Message
		}
	}
	return ""
}

func extractLastUserMessage(messages []SessionMessage) string {
	last := ""
	for _, m := range messages {
		if m.Sender == "user" {
			last = m.Message
		}
	}
	return last
}

func generateTitle(hobbyText string, itemsText string) string {
	titleMap := map[string]string{
		"ギター":   "アコギ入門セット",
		"コーヒー": "ハンドドリップコーヒーセット",
		"釣り":    "釣り入門スターターセット",
		"キャンプ": "キャンプスターターセット",
		"ヨガ":    "ヨガ入門セット",
		"イラスト": "イラスト道具セット",
		"ウクレレ": "ウクレレ入門セット",
		"カメラ":  "カメラスターターセット",
	}
	for key, title := range titleMap {
		if strings.Contains(hobbyText, key) {
			return title
		}
	}
	if hobbyText != "" {
		return fmt.Sprintf("%s スターターセット", hobbyText)
	}
	return "スターターセット一式"
}

func generateDescription(hobbyText string, _ string) string {
	descMap := map[string]string{
		"ギター":   "アコースティックギターを始めるのに必要な道具を一式揃えたセットです。チューニングからコードの練習まで、届いたその日から練習を始められます。初心者の方でも安心してスタートできる内容です。",
		"コーヒー": "ハンドドリップコーヒーを自宅で楽しむためのスターターセットです。器具をすべて揃えているので、届いたその日から本格的なコーヒーを淹れることができます。",
		"釣り":    "釣りを始めるために必要な基本道具を揃えたセットです。ロッドからリールまですぐに使える状態でお届けします。",
		"キャンプ": "ソロキャンプデビューに必要なアイテムを揃えたセットです。テント設営から調理まで、これだけで最初のキャンプを楽しめます。",
		"ヨガ":    "自宅でヨガを始めるためのスターターセットです。マットとサポートグッズが揃っているので、すぐに練習を開始できます。",
	}
	for key, desc := range descMap {
		if strings.Contains(hobbyText, key) {
			return desc
		}
	}
	if hobbyText != "" {
		return fmt.Sprintf("%sを始めるのに必要なものが揃ったセットです。丁寧に使用していたため状態は良好です。初心者の方でもすぐに始められる内容になっています。", hobbyText)
	}
	return "趣味を始めるのに必要なものが揃ったセットです。丁寧に使用していたため状態は良好です。"
}

func generateItems(hobbyText string) []ItemInput {
	itemMap := map[string][]ItemInput{
		"ギター": {
			{Name: "アコースティックギター", ConditionLabel: "good", IsEssential: true},
			{Name: "チューナー", ConditionLabel: "like_new", IsEssential: true},
			{Name: "ピック（5枚）", ConditionLabel: "new", IsEssential: true},
			{Name: "カポタスト", ConditionLabel: "like_new", IsEssential: false},
			{Name: "ギタースタンド", ConditionLabel: "good", IsEssential: false},
			{Name: "コード譜集", ConditionLabel: "good", IsEssential: false},
		},
		"コーヒー": {
			{Name: "ハンドドリッパー", ConditionLabel: "like_new", IsEssential: true},
			{Name: "コーヒーミル（手動）", ConditionLabel: "good", IsEssential: true},
			{Name: "細口ドリップポット", ConditionLabel: "good", IsEssential: true},
			{Name: "コーヒーサーバー", ConditionLabel: "like_new", IsEssential: false},
			{Name: "ペーパーフィルター（50枚）", ConditionLabel: "new", IsEssential: true},
			{Name: "デジタルスケール", ConditionLabel: "like_new", IsEssential: false},
		},
		"釣り": {
			{Name: "スピニングロッド", ConditionLabel: "good", IsEssential: true},
			{Name: "スピニングリール", ConditionLabel: "good", IsEssential: true},
			{Name: "釣り糸（ナイロン4号）", ConditionLabel: "new", IsEssential: true},
			{Name: "仕掛けセット", ConditionLabel: "new", IsEssential: true},
			{Name: "タックルボックス", ConditionLabel: "good", IsEssential: false},
		},
		"キャンプ": {
			{Name: "ソロテント（2人用）", ConditionLabel: "good", IsEssential: true},
			{Name: "シュラフ（3シーズン用）", ConditionLabel: "good", IsEssential: true},
			{Name: "マット", ConditionLabel: "good", IsEssential: true},
			{Name: "ガスバーナー", ConditionLabel: "like_new", IsEssential: true},
			{Name: "クッカーセット", ConditionLabel: "good", IsEssential: true},
			{Name: "ヘッドランプ", ConditionLabel: "like_new", IsEssential: false},
		},
	}

	if items, ok := itemMap[hobbyText]; ok {
		return items
	}

	return []ItemInput{
		{Name: "メインアイテム", ConditionLabel: "good", IsEssential: true},
		{Name: "サブアイテム", ConditionLabel: "good", IsEssential: false},
	}
}

func generateRecommended(hobbyText string) []RecommendedInput {
	recommendMap := map[string][]RecommendedInput{
		"ギター": {
			{Name: "ギター教則本", Importance: "recommended", Reason: "基礎から学べる入門書があると上達が早まります"},
			{Name: "替え弦", Importance: "required", Reason: "弦は消耗品です。切れた時のために予備を用意しておきましょう"},
			{Name: "指サック", Importance: "nice_to_have", Reason: "最初は指先が痛くなるので、慣れるまであると便利です"},
		},
		"コーヒー": {
			{Name: "コーヒー豆（200g）", Importance: "required", Reason: "まずはシングルオリジンの豆から試してみてください"},
			{Name: "温度計", Importance: "recommended", Reason: "お湯の温度管理で味が変わります"},
			{Name: "コーヒーノート", Importance: "nice_to_have", Reason: "レシピを記録すると自分好みの味を再現しやすくなります"},
		},
		"釣り": {
			{Name: "釣り用偏光グラス", Importance: "recommended", Reason: "水面の反射を抑えて魚の動きが見やすくなります"},
			{Name: "釣り用帽子", Importance: "recommended", Reason: "日差し対策と安全のために必要です"},
			{Name: "クーラーボックス", Importance: "nice_to_have", Reason: "釣れた魚を持ち帰る際に必要です"},
		},
		"キャンプ": {
			{Name: "焚き火台", Importance: "recommended", Reason: "焚き火はキャンプの醍醐味。直火禁止の場所でも使えます"},
			{Name: "ランタン", Importance: "required", Reason: "夜間の照明は必須です"},
			{Name: "折りたたみ椅子", Importance: "recommended", Reason: "快適なキャンプのためにあると便利です"},
		},
	}

	if items, ok := recommendMap[hobbyText]; ok {
		return items
	}

	return []RecommendedInput{
		{Name: "入門書", Importance: "recommended", Reason: "基礎知識を学ぶのに役立ちます"},
	}
}

func suggestedItemsForHobby(hobbyText string) []string {
	chipsMap := map[string][]string{
		"ギター":  {"ギター本体", "チューナー", "ピック", "カポタスト", "ギタースタンド", "コード譜"},
		"コーヒー": {"ドリッパー", "コーヒーミル", "ドリップポット", "コーヒーサーバー", "フィルター", "スケール"},
		"釣り":   {"ロッド", "リール", "釣り糸", "仕掛けセット", "タックルボックス", "ウキ"},
		"キャンプ": {"テント", "シュラフ", "マット", "バーナー", "クッカー", "ランタン"},
		"ヨガ":   {"ヨガマット", "ヨガブロック", "ヨガベルト", "ヨガウェア"},
		"イラスト": {"スケッチブック", "色鉛筆", "コピック", "水彩絵具", "筆"},
		"ウクレレ": {"ウクレレ本体", "チューナー", "ピック", "教則本"},
	}

	lowerHobby := strings.ToLower(hobbyText)
	for key, chips := range chipsMap {
		if strings.Contains(hobbyText, key) || strings.Contains(lowerHobby, strings.ToLower(key)) {
			return chips
		}
	}

	return []string{"メインアイテム", "サブアイテム", "収納ケース", "入門書"}
}
