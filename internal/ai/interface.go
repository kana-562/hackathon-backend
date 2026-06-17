package ai

import "hobby-relay-backend/internal/domain"

type ListingSupportResult struct {
	Message        string
	SuggestedChips []string
	Progress       domain.ProgressDTO
	Done           bool
	// Final result when Done=true
	Title             string
	Description       string
	HobbyText         string
	Items             []ItemInput
	RecommendedItems  []RecommendedInput
	BeginnerScore     int
	ReadinessScore    int
	PreviousOwnerNote string
	StartableSummary  string
}

type ItemInput struct {
	Name           string
	ConditionLabel string
	IsEssential    bool
}

type RecommendedInput struct {
	Name       string
	Importance string
	Reason     string
}

type SearchInterpretation struct {
	SmartMessage      string
	RelatedHobbies    []string
	MaxPrice          int
	MinBeginnerScore  int
	MinReadinessScore int
}

type Client interface {
	StartListingSupport(hobbyText string) (*ListingSupportResult, error)
	NextListingStep(sessionMessages []SessionMessage, userMessage string) (*ListingSupportResult, error)
	AnswerSetQuestion(setTitle string, items []string, userMessage string) (string, error)
	InterpretSearchQuery(query string) (*SearchInterpretation, error)
	GenerateStartPlan(setTitle string, hobbyName string) ([]domain.StartPlanStepDTO, error)
}

type SessionMessage struct {
	Sender  string
	Message string
}
