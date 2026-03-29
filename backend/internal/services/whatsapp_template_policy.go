package services

import (
	"strings"

	"github.com/fiando/propti/backend/internal/models"
)

type WhatsAppTemplateCategory string

const (
	WhatsAppTemplateCategoryUnknown        WhatsAppTemplateCategory = ""
	WhatsAppTemplateCategoryUtility        WhatsAppTemplateCategory = "utility"
	WhatsAppTemplateCategoryAuthentication WhatsAppTemplateCategory = "authentication"
)

type WhatsAppTemplateGoNoGoDecision string

const (
	WhatsAppTemplateGoNoGoDecisionGo   WhatsAppTemplateGoNoGoDecision = "go"
	WhatsAppTemplateGoNoGoDecisionNoGo WhatsAppTemplateGoNoGoDecision = "no-go"
)

type WhatsAppTemplateGoNoGoInputs struct {
	CostImpactHook       string `json:"costImpactHook,omitempty"`
	ConversionImpactHook string `json:"conversionImpactHook,omitempty"`
	RetentionImpactHook  string `json:"retentionImpactHook,omitempty"`
}

type WhatsAppTemplateDecisionInput struct {
	InCustomerServiceWindow bool
	RequestedMessageType    models.WhatsAppMessageType
	RequestedCategory       WhatsAppTemplateCategory
	IsCriticalTransactional bool
	UseCase                 string
	GoNoGoInputs            WhatsAppTemplateGoNoGoInputs
}

type WhatsAppTemplateDecisionOutput struct {
	Allowed                bool                           `json:"allowed"`
	GoNoGoDecision         WhatsAppTemplateGoNoGoDecision `json:"goNoGoDecision"`
	Reason                 string                         `json:"reason"`
	RecommendedMessageType models.WhatsAppMessageType     `json:"recommendedMessageType"`
	RecommendedCategory    WhatsAppTemplateCategory       `json:"recommendedCategory,omitempty"`
	CostImpact             bool                           `json:"costImpact"`
	ConversionImpact       bool                           `json:"conversionImpact"`
	RetentionImpact        bool                           `json:"retentionImpact"`
	GoNoGoInputs           WhatsAppTemplateGoNoGoInputs   `json:"goNoGoInputs"`
}

type WhatsAppTemplateDecisionPolicyOptions struct {
	EnableAuthenticationTemplates bool
}

type WhatsAppTemplateDecisionPolicy struct {
	enableAuthenticationTemplates bool
}

func NewWhatsAppTemplateDecisionPolicy(opts WhatsAppTemplateDecisionPolicyOptions) *WhatsAppTemplateDecisionPolicy {
	return &WhatsAppTemplateDecisionPolicy{enableAuthenticationTemplates: opts.EnableAuthenticationTemplates}
}

func (p *WhatsAppTemplateDecisionPolicy) Decide(input WhatsAppTemplateDecisionInput) WhatsAppTemplateDecisionOutput {
	decision := WhatsAppTemplateDecisionOutput{
		Allowed:                false,
		GoNoGoDecision:         WhatsAppTemplateGoNoGoDecisionNoGo,
		RecommendedMessageType: models.WhatsAppMessageTypeText,
		GoNoGoInputs:           input.GoNoGoInputs,
		CostImpact:             strings.TrimSpace(input.GoNoGoInputs.CostImpactHook) != "",
		ConversionImpact:       strings.TrimSpace(input.GoNoGoInputs.ConversionImpactHook) != "",
		RetentionImpact:        strings.TrimSpace(input.GoNoGoInputs.RetentionImpactHook) != "",
	}

	if input.InCustomerServiceWindow {
		if input.RequestedMessageType != models.WhatsAppMessageTypeTemplate {
			decision.Allowed = true
			decision.GoNoGoDecision = WhatsAppTemplateGoNoGoDecisionGo
			decision.Reason = "customer-service window is open; free-form messaging is required"
			return decision
		}
		decision.Reason = "customer-service window is open; templates should not be used"
		return decision
	}

	decision.RecommendedMessageType = models.WhatsAppMessageTypeTemplate
	decision.RecommendedCategory = WhatsAppTemplateCategoryUtility

	if input.RequestedMessageType != models.WhatsAppMessageTypeTemplate {
		decision.Reason = "customer-service window is closed; only utility templates for critical transactional use cases are allowed"
		return decision
	}

	if input.RequestedCategory == WhatsAppTemplateCategoryAuthentication {
		decision.RecommendedCategory = WhatsAppTemplateCategoryAuthentication
		if !p.enableAuthenticationTemplates {
			decision.Reason = "authentication templates are disabled by default; enable explicit feature flag to allow"
			return decision
		}
		decision.Allowed = input.IsCriticalTransactional
		if input.IsCriticalTransactional {
			decision.GoNoGoDecision = WhatsAppTemplateGoNoGoDecisionGo
			decision.Reason = "authentication template allowed via feature flag for critical transactional flow"
		} else {
			decision.Reason = "authentication template requires critical transactional use case"
		}
		return decision
	}

	if input.RequestedCategory != WhatsAppTemplateCategoryUtility {
		decision.Reason = "only utility templates are allowed outside customer-service window"
		return decision
	}
	if !input.IsCriticalTransactional {
		decision.Reason = "utility templates outside customer-service window require critical transactional use case"
		return decision
	}

	decision.Allowed = true
	decision.GoNoGoDecision = WhatsAppTemplateGoNoGoDecisionGo
	decision.Reason = "utility template approved for critical transactional use case outside customer-service window"
	decision.RecommendedCategory = WhatsAppTemplateCategoryUtility
	return decision
}
