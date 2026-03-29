package services

import (
	"testing"

	"github.com/fiando/propti/backend/internal/models"
)

func TestWhatsAppTemplatePolicyInWindowAllowsOnlyFreeForm(t *testing.T) {
	policy := NewWhatsAppTemplateDecisionPolicy(WhatsAppTemplateDecisionPolicyOptions{})

	decision := policy.Decide(WhatsAppTemplateDecisionInput{
		InCustomerServiceWindow: true,
		RequestedMessageType:    models.WhatsAppMessageTypeText,
		UseCase:                 "follow-up response",
		GoNoGoInputs: WhatsAppTemplateGoNoGoInputs{
			CostImpactHook:       "no-template-cost",
			ConversionImpactHook: "high-intent-user",
			RetentionImpactHook:  "conversation-active",
		},
	})
	if decision.GoNoGoDecision != WhatsAppTemplateGoNoGoDecisionGo {
		t.Fatalf("expected go decision for in-window free-form message, got %q", decision.GoNoGoDecision)
	}
	if !decision.Allowed {
		t.Fatal("expected in-window free-form message to be allowed")
	}

	templateDecision := policy.Decide(WhatsAppTemplateDecisionInput{
		InCustomerServiceWindow: true,
		RequestedMessageType:    models.WhatsAppMessageTypeTemplate,
		RequestedCategory:       WhatsAppTemplateCategoryUtility,
		IsCriticalTransactional: true,
	})
	if templateDecision.GoNoGoDecision != WhatsAppTemplateGoNoGoDecisionNoGo {
		t.Fatalf("expected no-go decision for in-window template, got %q", templateDecision.GoNoGoDecision)
	}
	if templateDecision.Allowed {
		t.Fatal("expected in-window template message to be blocked")
	}
	if templateDecision.RecommendedMessageType != models.WhatsAppMessageTypeText {
		t.Fatalf("expected free-form recommendation, got %q", templateDecision.RecommendedMessageType)
	}
}

func TestWhatsAppTemplatePolicyOutOfWindowAllowsOnlyCriticalUtilityTemplate(t *testing.T) {
	policy := NewWhatsAppTemplateDecisionPolicy(WhatsAppTemplateDecisionPolicyOptions{})

	blocked := policy.Decide(WhatsAppTemplateDecisionInput{
		InCustomerServiceWindow: false,
		RequestedMessageType:    models.WhatsAppMessageTypeTemplate,
		RequestedCategory:       WhatsAppTemplateCategoryUtility,
		IsCriticalTransactional: false,
		UseCase:                 "marketing drip",
	})
	if blocked.Allowed {
		t.Fatal("expected non-critical utility template to be blocked out of window")
	}
	if blocked.GoNoGoDecision != WhatsAppTemplateGoNoGoDecisionNoGo {
		t.Fatalf("expected no-go for non-critical utility template, got %q", blocked.GoNoGoDecision)
	}

	allowed := policy.Decide(WhatsAppTemplateDecisionInput{
		InCustomerServiceWindow: false,
		RequestedMessageType:    models.WhatsAppMessageTypeTemplate,
		RequestedCategory:       WhatsAppTemplateCategoryUtility,
		IsCriticalTransactional: true,
		UseCase:                 "booking confirmation",
	})
	if !allowed.Allowed {
		t.Fatal("expected critical utility template to be allowed out of window")
	}
	if allowed.GoNoGoDecision != WhatsAppTemplateGoNoGoDecisionGo {
		t.Fatalf("expected go for critical utility template, got %q", allowed.GoNoGoDecision)
	}
	if allowed.RecommendedMessageType != models.WhatsAppMessageTypeTemplate {
		t.Fatalf("expected template recommendation, got %q", allowed.RecommendedMessageType)
	}
	if allowed.RecommendedCategory != WhatsAppTemplateCategoryUtility {
		t.Fatalf("expected utility recommendation, got %q", allowed.RecommendedCategory)
	}
}

func TestWhatsAppTemplatePolicyAuthenticationTemplatesAreFeatureFlagged(t *testing.T) {
	policy := NewWhatsAppTemplateDecisionPolicy(WhatsAppTemplateDecisionPolicyOptions{})

	blocked := policy.Decide(WhatsAppTemplateDecisionInput{
		InCustomerServiceWindow: false,
		RequestedMessageType:    models.WhatsAppMessageTypeTemplate,
		RequestedCategory:       WhatsAppTemplateCategoryAuthentication,
		IsCriticalTransactional: true,
		UseCase:                 "otp verification",
	})
	if blocked.Allowed {
		t.Fatal("expected authentication template to be blocked by default")
	}
	if blocked.GoNoGoDecision != WhatsAppTemplateGoNoGoDecisionNoGo {
		t.Fatalf("expected no-go for auth template when feature disabled, got %q", blocked.GoNoGoDecision)
	}

	enabledPolicy := NewWhatsAppTemplateDecisionPolicy(WhatsAppTemplateDecisionPolicyOptions{EnableAuthenticationTemplates: true})
	allowed := enabledPolicy.Decide(WhatsAppTemplateDecisionInput{
		InCustomerServiceWindow: false,
		RequestedMessageType:    models.WhatsAppMessageTypeTemplate,
		RequestedCategory:       WhatsAppTemplateCategoryAuthentication,
		IsCriticalTransactional: true,
		UseCase:                 "otp verification",
	})
	if !allowed.Allowed {
		t.Fatal("expected authentication template to be allowed when feature is enabled")
	}
	if allowed.RecommendedCategory != WhatsAppTemplateCategoryAuthentication {
		t.Fatalf("expected authentication recommendation, got %q", allowed.RecommendedCategory)
	}
}

func TestWhatsAppTemplatePolicyCarriesGoNoGoImpactInputs(t *testing.T) {
	policy := NewWhatsAppTemplateDecisionPolicy(WhatsAppTemplateDecisionPolicyOptions{})

	decision := policy.Decide(WhatsAppTemplateDecisionInput{
		InCustomerServiceWindow: false,
		RequestedMessageType:    models.WhatsAppMessageTypeTemplate,
		RequestedCategory:       WhatsAppTemplateCategoryUtility,
		IsCriticalTransactional: true,
		GoNoGoInputs: WhatsAppTemplateGoNoGoInputs{
			CostImpactHook:       "paid-conversation-category",
			ConversionImpactHook: "recover-transaction",
			RetentionImpactHook:  "protect-user-trust",
		},
	})

	if decision.GoNoGoInputs.CostImpactHook != "paid-conversation-category" {
		t.Fatalf("expected cost impact hook to be preserved, got %q", decision.GoNoGoInputs.CostImpactHook)
	}
	if decision.GoNoGoInputs.ConversionImpactHook != "recover-transaction" {
		t.Fatalf("expected conversion impact hook to be preserved, got %q", decision.GoNoGoInputs.ConversionImpactHook)
	}
	if decision.GoNoGoInputs.RetentionImpactHook != "protect-user-trust" {
		t.Fatalf("expected retention impact hook to be preserved, got %q", decision.GoNoGoInputs.RetentionImpactHook)
	}
}
