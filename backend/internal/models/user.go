package models

import "time"

type UserRole string
type SubscriptionTier string
type SubscriptionStatus string

const (
	UserRoleBuyer  UserRole = "buyer"
	UserRoleSeller UserRole = "seller"
	UserRoleBoth   UserRole = "both"

	SubscriptionFree    SubscriptionTier = "free"
	SubscriptionBasic   SubscriptionTier = "basic"
	SubscriptionPremium SubscriptionTier = "premium"
	SubscriptionPro     SubscriptionTier = "pro"

	SubscriptionActive       SubscriptionStatus = "active"
	SubscriptionExpiringSoon SubscriptionStatus = "expiring_soon"
	SubscriptionExpired      SubscriptionStatus = "expired"
)

type UserPreferences struct {
	FavoriteLocations []string `json:"favoriteLocations" dynamodbav:"favoriteLocations"`
	SearchHistory     []string `json:"searchHistory" dynamodbav:"searchHistory"`
	Notifications     bool     `json:"notifications" dynamodbav:"notifications"`
}

type Subscription struct {
	Tier                SubscriptionTier `json:"tier" dynamodbav:"tier"`
	MonthlyListingsUsed int              `json:"monthlyListingsUsed" dynamodbav:"monthlyListingsUsed"`
	VoiceUsageMonth     string           `json:"voiceUsageMonth,omitempty" dynamodbav:"voiceUsageMonth,omitempty"`
	VoiceSecondsUsed    int              `json:"voiceSecondsUsed" dynamodbav:"voiceSecondsUsed"`
	ActiveListingsCount int              `json:"activeListingsCount" dynamodbav:"-"`
	RenewDate           *time.Time       `json:"renewDate,omitempty" dynamodbav:"renewDate,omitempty"`
	PaymentCustomerID   string           `json:"paymentCustomerId,omitempty" dynamodbav:"paymentCustomerId,omitempty"`
}

type ContactRevealThrottle struct {
	WindowStartedAt time.Time `json:"windowStartedAt,omitempty" dynamodbav:"windowStartedAt,omitempty"`
	RevealCount     int       `json:"revealCount" dynamodbav:"revealCount"`
}

type User struct {
	PK                    string                `json:"pk" dynamodbav:"PK"`
	SK                    string                `json:"sk" dynamodbav:"SK"`
	UserID                string                `json:"userId" dynamodbav:"userId"`
	GoogleID              string                `json:"googleId" dynamodbav:"googleId"`
	Email                 string                `json:"email" dynamodbav:"email"`
	Name                  string                `json:"name" dynamodbav:"name"`
	ProfilePicture        string                `json:"profilePicture" dynamodbav:"profilePicture"`
	Phone                 string                `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	Role                  UserRole              `json:"role" dynamodbav:"role"`
	Preferences           UserPreferences       `json:"preferences" dynamodbav:"preferences"`
	SavedListingIDs       []string              `json:"savedListingIds,omitempty" dynamodbav:"savedListingIds,omitempty"`
	Subscription          Subscription          `json:"subscription" dynamodbav:"subscription"`
	ContactRevealThrottle ContactRevealThrottle `json:"contactRevealThrottle,omitempty" dynamodbav:"contactRevealThrottle,omitempty"`
	WhatsAppLinkedPhone   string                `json:"whatsAppLinkedPhone,omitempty" dynamodbav:"whatsAppLinkedPhone,omitempty"`
	WhatsAppLinkedAt      *time.Time            `json:"whatsAppLinkedAt,omitempty" dynamodbav:"whatsAppLinkedAt,omitempty"`
	WhatsAppVerifiedAt    *time.Time            `json:"whatsAppVerifiedAt,omitempty" dynamodbav:"whatsAppVerifiedAt,omitempty"`
	CreatedAt             time.Time             `json:"createdAt" dynamodbav:"createdAt"`
	LastLoginAt           time.Time             `json:"lastLoginAt" dynamodbav:"lastLoginAt"`
}

type GoogleAuthRequest struct {
	IDToken     string `json:"idToken,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
}

type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	User        *User  `json:"user"`
}

type UpdateUserRequest struct {
	Phone       *string          `json:"phone,omitempty"`
	Role        *UserRole        `json:"role,omitempty"`
	Preferences *UserPreferences `json:"preferences,omitempty"`
}

func (u *User) IsWhatsAppWriteEligible() bool {
	if u == nil {
		return false
	}
	if u.WhatsAppVerifiedAt == nil {
		return false
	}
	return u.WhatsAppLinkedPhone != ""
}
