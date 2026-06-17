package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"hobby-relay-backend/internal/domain"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(apiKey, model string) Client {
	if model == "" {
		model = openai.GPT4oMini
	}
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (c *OpenAIClient) chat(messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := c.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   1024,
		Temperature: 0.7,
	})
	if err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) chatJSON(messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := c.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		MaxTokens:   1024,
		Temperature: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) StartListingSupport(hobbyText string) (*ListingSupportResult, error) {
	systemPrompt := `あなたはフリマアプリ「ホビーリレー」の出品サポートAIです。
ユーザーが使わなくなった趣味道具を出品するのを手伝います。
返答は必ず日本語で、短く親しみやすいトーンで話してください。`

	userMsg := fmt.Sprintf("「%s」のセットを出品したいです。", hobbyText)

	msg, err := c.chat([]openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
		{Role: openai.ChatMessageRoleUser, Content: userMsg},
		{Role: openai.ChatMessageRoleSystem, Content: fmt.Sprintf(`
%sのセットですね。入っているものを教えてください。
以下のような形式で返答してください（1〜2文で）。
チップとして提案するアイテム名も5〜7個提案してください（%sに関連するもの）。`, hobbyText, hobbyText)},
	})
	if err != nil {
		return NewMockClient().StartListingSupport(hobbyText)
	}

	chips := defaultChipsForHobby(hobbyText)

	return &ListingSupportResult{
		Message:        msg,
		SuggestedChips: chips,
		Progress:       domain.ProgressDTO{Current: 1, Total: 5},
		Done:           false,
	}, nil
}

func (c *OpenAIClient) NextListingStep(sessionMessages []SessionMessage, userMessage string) (*ListingSupportResult, error) {
	step := len(sessionMessages)/2 + 1

	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `あなたはフリマアプリ「ホビーリレー」の出品サポートAIです。
ステップ1:入っているもの確認 → ステップ2:状態確認 → ステップ3:価格希望 → ステップ4:前オーナーの一言 → ステップ5:出品内容作成
返答は日本語で短く。`,
		},
	}
	for _, m := range sessionMessages {
		role := openai.ChatMessageRoleUser
		if m.Sender == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{Role: role, Content: m.Message})
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: userMessage})

	if step >= 5 {
		return c.finalizeListingSupport(sessionMessages, userMessage)
	}

	var nextPrompt string
	var chips []string

	switch step {
	case 2:
		nextPrompt = "アイテムの状態を教えてください。"
		chips = []string{"ほぼ新品", "目立った傷なし", "やや傷あり", "動作未確認"}
	case 3:
		nextPrompt = "希望価格を教えてください。"
		chips = []string{"3,000円", "5,000円", "8,000円", "10,000円", "15,000円以上"}
	case 4:
		nextPrompt = "最後に、前の持ち主からひと言メッセージをどうぞ。"
		chips = []string{}
	default:
		nextPrompt = "ありがとうございます。もう少し教えてください。"
		chips = []string{}
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("次のステップに進んでください: %s 短く1〜2文で。", nextPrompt),
	})

	msg, err := c.chat(messages)
	if err != nil {
		msg = nextPrompt
	}

	return &ListingSupportResult{
		Message:        msg,
		SuggestedChips: chips,
		Progress:       domain.ProgressDTO{Current: step, Total: 5},
		Done:           false,
	}, nil
}

func (c *OpenAIClient) finalizeListingSupport(sessionMessages []SessionMessage, lastMessage string) (*ListingSupportResult, error) {
	history := ""
	for _, m := range sessionMessages {
		history += fmt.Sprintf("%s: %s\n", m.Sender, m.Message)
	}
	history += fmt.Sprintf("user: %s\n", lastMessage)

	prompt := fmt.Sprintf(`以下の出品サポートの会話から出品情報を抽出してJSONを返してください。

会話:
%s

以下のJSON形式で返してください:
{
  "title": "セットタイトル（〜セットのような形式）",
  "description": "説明文（2〜3文）",
  "hobbyText": "趣味名",
  "beginnerScore": 3,
  "readinessScore": 80,
  "startableSummary": "このセットで始められること（1文）",
  "previousOwnerNote": "前の持ち主の一言",
  "items": [
    {"name": "アイテム名", "conditionLabel": "like_new", "isEssential": true}
  ],
  "recommendedItems": [
    {"name": "推奨アイテム名", "importance": "recommended", "reason": "理由"}
  ]
}

conditionLabelは: new, like_new, good, fair, unknown のいずれか
importanceは: required, recommended, nice_to_have のいずれか`, history)

	raw, err := c.chatJSON([]openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	})
	if err != nil {
		return NewMockClient().NextListingStep(sessionMessages, lastMessage)
	}

	var result struct {
		Title             string `json:"title"`
		Description       string `json:"description"`
		HobbyText         string `json:"hobbyText"`
		BeginnerScore     int    `json:"beginnerScore"`
		ReadinessScore    int    `json:"readinessScore"`
		StartableSummary  string `json:"startableSummary"`
		PreviousOwnerNote string `json:"previousOwnerNote"`
		Items             []struct {
			Name           string `json:"name"`
			ConditionLabel string `json:"conditionLabel"`
			IsEssential    bool   `json:"isEssential"`
		} `json:"items"`
		RecommendedItems []struct {
			Name       string `json:"name"`
			Importance string `json:"importance"`
			Reason     string `json:"reason"`
		} `json:"recommendedItems"`
	}

	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return NewMockClient().NextListingStep(sessionMessages, lastMessage)
	}

	items := make([]ItemInput, len(result.Items))
	for i, it := range result.Items {
		items[i] = ItemInput{Name: it.Name, ConditionLabel: it.ConditionLabel, IsEssential: it.IsEssential}
	}
	recs := make([]RecommendedInput, len(result.RecommendedItems))
	for i, r := range result.RecommendedItems {
		recs[i] = RecommendedInput{Name: r.Name, Importance: r.Importance, Reason: r.Reason}
	}

	return &ListingSupportResult{
		Message:           "出品内容を作成しました。確認画面へ進みます。",
		SuggestedChips:    []string{},
		Progress:          domain.ProgressDTO{Current: 5, Total: 5},
		Done:              true,
		Title:             result.Title,
		Description:       result.Description,
		HobbyText:         result.HobbyText,
		BeginnerScore:     result.BeginnerScore,
		ReadinessScore:    result.ReadinessScore,
		StartableSummary:  result.StartableSummary,
		PreviousOwnerNote: result.PreviousOwnerNote,
		Items:             items,
		RecommendedItems:  recs,
	}, nil
}

func (c *OpenAIClient) AnswerSetQuestion(setTitle string, items []string, userMessage string) (string, error) {
	itemList := strings.Join(items, "、")
	msg, err := c.chat([]openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(`あなたはフリマアプリの商品「%s」についての質問に答えるAIです。
このセットには: %s が含まれています。
短く（2〜3文以内）、親しみやすいトーンで日本語で答えてください。`, setTitle, itemList),
		},
		{Role: openai.ChatMessageRoleUser, Content: userMessage},
	})
	if err != nil {
		return NewMockClient().AnswerSetQuestion(setTitle, items, userMessage)
	}
	return msg, nil
}

func (c *OpenAIClient) InterpretSearchQuery(query string) (*SearchInterpretation, error) {
	prompt := fmt.Sprintf(`フリマアプリの検索クエリを解釈してJSONを返してください。

クエリ: 「%s」

以下のJSON形式で返してください:
{
  "smartMessage": "ユーザーへの補足メッセージ（〜に近いセットを表示しています、のような形式）",
  "relatedHobbies": ["関連する趣味名1", "関連する趣味名2"],
  "maxPrice": 0,
  "minBeginnerScore": 0,
  "minReadinessScore": 0
}

趣味名はこの中から選んでください: ギター, ウクレレ, キーボード, 筋トレ, ヨガ, キャンプ, 釣り, カメラ, イラスト, コーヒー, 自炊, 語学, 資格勉強
maxPriceは「1万円以内」「10000円以内」等から数値を抽出（なければ0）
minBeginnerScoreは「初心者向け」等から推測（なければ0、最大5）
relatedHobbiesはクエリに関連する趣味（なければ空配列）`, query)

	raw, err := c.chatJSON([]openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	})
	if err != nil {
		return NewMockClient().InterpretSearchQuery(query)
	}

	var result struct {
		SmartMessage      string   `json:"smartMessage"`
		RelatedHobbies    []string `json:"relatedHobbies"`
		MaxPrice          int      `json:"maxPrice"`
		MinBeginnerScore  int      `json:"minBeginnerScore"`
		MinReadinessScore int      `json:"minReadinessScore"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return NewMockClient().InterpretSearchQuery(query)
	}

	return &SearchInterpretation{
		SmartMessage:      result.SmartMessage,
		RelatedHobbies:    result.RelatedHobbies,
		MaxPrice:          result.MaxPrice,
		MinBeginnerScore:  result.MinBeginnerScore,
		MinReadinessScore: result.MinReadinessScore,
	}, nil
}

func (c *OpenAIClient) GenerateStartPlan(setTitle string, hobbyName string) ([]domain.StartPlanStepDTO, error) {
	prompt := fmt.Sprintf(`フリマアプリで「%s」（%s）を購入したユーザー向けに7日間スタートプランをJSONで作ってください。

以下のJSON形式で返してください:
{
  "steps": [
    {"dayNo": 1, "title": "Day 1のタイトル", "body": "具体的な行動（1〜2文）"},
    {"dayNo": 2, "title": "Day 2のタイトル", "body": "具体的な行動"},
    ...
    {"dayNo": 7, "title": "Day 7のタイトル", "body": "具体的な行動"}
  ]
}

初心者が実際に始められる具体的な内容にしてください。日本語で。`, setTitle, hobbyName)

	raw, err := c.chatJSON([]openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	})
	if err != nil {
		return NewMockClient().GenerateStartPlan(setTitle, hobbyName)
	}

	var result struct {
		Steps []struct {
			DayNo int    `json:"dayNo"`
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"steps"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return NewMockClient().GenerateStartPlan(setTitle, hobbyName)
	}

	steps := make([]domain.StartPlanStepDTO, len(result.Steps))
	for i, s := range result.Steps {
		steps[i] = domain.StartPlanStepDTO{DayNo: s.DayNo, Title: s.Title, Body: s.Body}
	}
	return steps, nil
}

// defaultChipsForHobby returns suggested item chips based on hobby text.
func defaultChipsForHobby(hobbyText string) []string {
	lower := strings.ToLower(hobbyText)
	switch {
	case strings.Contains(lower, "ギター") || strings.Contains(lower, "guitar"):
		return []string{"ギター本体", "チューナー", "ピック", "カポ", "教本", "替え弦", "スタンド"}
	case strings.Contains(lower, "キャンプ") || strings.Contains(lower, "camp"):
		return []string{"テント", "チェア", "ランタン", "寝袋", "クッカー", "焚き火台"}
	case strings.Contains(lower, "コーヒー") || strings.Contains(lower, "coffee"):
		return []string{"コーヒーミル", "ドリッパー", "サーバー", "ケトル", "スケール", "豆"}
	case strings.Contains(lower, "カメラ") || strings.Contains(lower, "camera"):
		return []string{"カメラ本体", "レンズ", "SDカード", "バッテリー", "カメラバッグ"}
	case strings.Contains(lower, "ウクレレ") || strings.Contains(lower, "ukulele"):
		return []string{"ウクレレ本体", "チューナー", "教本", "ピック", "ケース"}
	case strings.Contains(lower, "ヨガ") || strings.Contains(lower, "yoga"):
		return []string{"ヨガマット", "ブロック", "ストラップ", "ボルスター", "DVD"}
	case strings.Contains(lower, "イラスト") || strings.Contains(lower, "illust"):
		return []string{"ペンタブ", "コピック", "スケッチブック", "入門書", "替え芯"}
	default:
		// Parse price strings like "3,000円" to integers for price chips
		_ = strconv.Itoa(0)
		return []string{"メイン道具", "サブ道具", "入門書", "ケース・バッグ", "消耗品"}
	}
}
