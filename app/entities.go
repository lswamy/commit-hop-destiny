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
	Values      DestinyHistoricalStatsValue
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
