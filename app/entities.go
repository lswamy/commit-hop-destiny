package app

import "encoding/json"

type ConfigResponse struct {
	Response ConfigDestinyManifest
}

type ConfigDestinyManifest struct {
	Version                      string
	MobileAssetContentPath       string
	MobileGearAssetDataBase      []GearAssetDataBaseDefinition
	JsonWorldContentPaths        map[string]string
	MobileClanBannerDatabasePath string
}

type GearAssetDataBaseDefinition struct {
	Version int32
	Path    string
}

type ActivityResponse struct {
	Response DestinyActivityHistoryResults
}

type PGCRResponse struct {
	Response DestinyPostGameCarnageReportData
}

type CharacterResponse struct {
	Response DestinyCharacterResponse
}

type ProfileResponse struct {
	Response DestinyProfileResponse
}

type DestinyActivityHistoryResults struct {
	Activities []DestinyHistoricalStatsPeriodGroup
}

type DestinyHistoricalStatsPeriodGroup struct {
	Period          string
	ActivityDetails DestinyHistoricalStatsActivity
	Values          DestinyHistoricalStatsValue
}

type DestinyHistoricalStatsActivity struct {
	ReferenceId          uint32
	DirectorActivityHash uint32
	InstanceId           json.Number
	Mode                 int32
	Modes                []int32
	IsPrivate            bool
	MembershipType       int32
}

type DestinyHistoricalStatsValuePair struct {
	Value        float32
	DisplayValue string
}

type DestinyHistoricalStatsValue struct {
	StatId     string
	ActivityId int64
	Basic      DestinyHistoricalStatsValuePair
	Pga        DestinyHistoricalStatsValuePair
	Weighted   DestinyHistoricalStatsValuePair
}

type DestinyPostGameCarnageReportData struct {
	Period          string
	ActivityDetails DestinyHistoricalStatsActivity
	Entries         []DestinyPostGameCarnageReportEntry
	Teams           []DestinyPostGameCarnageReportTeamEntry
}

type DestinyPostGameCarnageReportEntry struct {
	Standing    int32
	Score       DestinyHistoricalStatsValue
	Player      DestinyPlayer
	CharacterId json.Number
	Values      map[string]DestinyHistoricalStatsValue
	Extended    DestinyPostGameCarnageReportExtendedData
}

type DestinyPostGameCarnageReportTeamEntry struct {
	TeamId   int32
	Standing DestinyHistoricalStatsValue
	Score    DestinyHistoricalStatsValue
	TeamName string
}

type DestinyPostGameCarnageReportExtendedData struct {
	Weapons []DestinyHistoricalWeaponStats
	Values  DestinyHistoricalStatsValue
}

type DestinyHistoricalWeaponStats struct {
	ReferenceId json.Number
	Values      map[string]DestinyHistoricalStatsValue
}

type DestinyPlayer struct {
	DestinyUserInfo UserInfoCard
	CharacterClass  string
	CharacterLevel  int32
	LightLevel      int32
	ClanName        string
	ClanTag         string
}

type UserInfoCard struct {
	SupplementalDisplayName string
	IconPath                string
	IsPublic                bool
	MembershipType          int32
	MembershipId            string
	DisplayName             string
	BungieGlobalDisplayName string
}

type DisplayPropertiesDefinition struct {
	Description string
	Name        string
	Icon        string
	HasIcon     bool
}

type ActivityDefinition struct {
	DisplayProperties DisplayPropertiesDefinition
	DestinationHash   uint32
	PlaceHash         uint32
	ActivityTypeHash  uint32
	PcgrImage         string
}

type InventoryItemDefinition struct {
	DisplayProperties          DisplayPropertiesDefinition
	Screenshot                 string
	ItemTypeDisplayName        string
	ItemTypeAndTierDisplayName string
	ItemType                   int32
	ItemSubType                int32
	ClassType                  int32
}

type DestinyCharacterResponse struct {
	Character SingleComponentResponseOfDestinyCharacterComponent	
}

type SingleComponentResponseOfDestinyCharacterComponent struct {
	Data DestinyCharacterComponent
	Privacy int32
	Disabled bool
}

type DestinyColor struct {
	Red byte
	Green byte
	Blue byte
	Alpha byte
}

type DestinyCharacterComponent struct {
	MembershipId json.Number
	MembershipType int32
	CharacterId json.Number
	DateLastPlayed string
	MinutesPlayedThisSession json.Number
	MinutesPlayedTotal json.Number
	Light int32
	Stats map[uint32]int32
	RaceHash uint32
	GenderHash uint32
	ClassHash uint32
	RaceType int32
	ClassType int32
	GenderType int32
	EmblemPath string
	EmblemBackgroundPath string
	EmblemHash uint32
	EmblemColor DestinyColor
	// LevelProgression string
	BaseCharacterLevel int32
	PercentToNextLevel float32
	TitleRecordHash uint32
}

type DestinyProfileResponse struct {
	Profile SingleComponentResponseOfDestinyProfileComponent
}

type SingleComponentResponseOfDestinyProfileComponent struct {
	Data DestinyProfileComponent
	Privacy int32
	Disabled bool
}

type DestinyProfileComponent struct {
	UserInfo UserInfoCard
	DateLastPlayed string
	CharacterIds []json.Number
}
