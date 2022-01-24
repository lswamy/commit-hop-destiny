package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/lswamy/commit-hop-destiny/app"
)

const HttpPort = 8080

const ApiRootPath string = "https://www.bungie.net"
const DestinyMembershipType int = 2
const ActivityModeTypePve int = 7
const ActivityModeTypePvp int = 5
const NumActivities int = 5

var cwd, _ = os.Getwd()
var worldContent *app.SQLiteDB

type PageContent struct {
	Name            string
	MainCharacterId int64
	Activities      []ActivityData
	WeaponKills     map[int64]float32
	WeaponTypeKills map[string]float32
	WeaponDefs      map[int64]app.InventoryItemDefinition
	WeaponTypeNames map[string]string
	CharacterStats  struct {
		WeaponKills map[int64]float32
	}
	JsonData map[string]string
}

type ProfileContent struct {
	Data app.DestinyProfileResponse
	CharacterData app.DestinyCharacterResponse
	Color []int
}

type ActivityData struct {
	Definition          app.ActivityDefinition
	ReferenceDefinition app.ActivityDefinition
	InstanceId          int64
	Period              string
	PgcrEntries         []app.DestinyPostGameCarnageReportEntry
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	destinyCharacterId := viper.GetInt("destiny_character_id")
	destinyMembershipId := viper.GetInt("destiny_membership_id")
	
	content := ProfileContent{}

	profileComponents := []string{"Profile"}
	charComponents := []string{"Characters"}
	
	fmt.Println("requesting profile...")
	content.Data = requestProfile(int32(DestinyMembershipType), int64(destinyMembershipId), profileComponents)

	fmt.Println("requesting character...")
	content.CharacterData = requestCharacter(int32(DestinyMembershipType), int64(destinyMembershipId), int64(destinyCharacterId), charComponents)
	
	// s, _ := json.MarshalIndent(content.CharacterData.Character.Data.EmblemColor, "", "\t")
	// fmt.Print(string(s))

	content.Color = rgbContrast(
		content.CharacterData.Character.Data.EmblemColor.Red,
		content.CharacterData.Character.Data.EmblemColor.Green,
		content.CharacterData.Character.Data.EmblemColor.Blue,
		content.CharacterData.Character.Data.EmblemColor.Alpha,
	)

	tplPath := cwd + "/web/tmpl/index.html"
	tpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"numToStr": func(num json.Number) string {
			return num.String()
		},
		"joinInts": func(nums []int, delim string) string {
			return strings.Trim(strings.Join(strings.Split(fmt.Sprint(nums), " "), delim), "[]")
		},
	}).ParseFiles(tplPath))
	if tplErr := tpl.Execute(w, content); tplErr != nil {
		fmt.Println(("exec err"))
		log.Fatal(tplErr)
	}
}

func activitiesHandler(w http.ResponseWriter, r *http.Request) {
	destinyCharacterId := viper.GetInt("destiny_character_id")
	characterActivities := requestActivityHistory(DestinyMembershipType, viper.GetInt("destiny_membership_id"), destinyCharacterId, ActivityModeTypePvp, NumActivities)

	pageContent := PageContent{}
	pageContent.MainCharacterId = int64(destinyCharacterId)
	pageContent.Activities = make([]ActivityData, 0)
	pageContent.WeaponKills = make(map[int64]float32)
	pageContent.WeaponTypeKills = make(map[string]float32)
	pageContent.CharacterStats.WeaponKills = make(map[int64]float32)
	pageContent.JsonData = make(map[string]string)

	// s, _ := json.MarshalIndent(pageContent, "", "\t")
	// fmt.Print(string(s))

	for a := range characterActivities.Activities {
		ad := ActivityData{}
		activity := characterActivities.Activities[a]
		activityHash := int32(activity.ActivityDetails.DirectorActivityHash)
		referenceId := int32(activity.ActivityDetails.ReferenceId)
		instanceId, _ := activity.ActivityDetails.InstanceId.Int64()

		activityDef := worldContent.GetActivityDefinition(activityHash)
		// s, _ := json.MarshalIndent(activityDef, "", "\t")
		// fmt.Print(string(s))

		if activityHash != referenceId {
			ad.ReferenceDefinition = worldContent.GetActivityDefinition(referenceId)
		}

		ad.InstanceId = instanceId
		ad.Definition = activityDef
		
		activityTime, _ := time.Parse(time.RFC3339, activity.Period)
		fmt.Println(activityTime)
		ad.Period = activityTime.Format("Jan 2, 2006")

		pgcr := requestPostGameCarnageReport(instanceId)
		ad.PgcrEntries = pgcr.Entries

		for _, entry := range pgcr.Entries {
			entryCharId, _ := entry.CharacterId.Int64()
			for _, weapon := range entry.Extended.Weapons {
				weaponId, _ := weapon.ReferenceId.Int64()
				uniqueKills := weapon.Values["uniqueWeaponKills"].Basic.Value
				pageContent.WeaponKills[weaponId] += uniqueKills

				if entryCharId == int64(destinyCharacterId) {
					pageContent.CharacterStats.WeaponKills[weaponId] += uniqueKills
				}
			}
		}

		pageContent.Activities = append(pageContent.Activities, ad)
		allWeaponsJson, _ := json.Marshal(pageContent.WeaponKills)
		pageContent.JsonData["allWeapons"] = string(allWeaponsJson)
	}

	pageContent.WeaponDefs = make(map[int64]app.InventoryItemDefinition)
	for weaponId := range pageContent.WeaponKills {
		def := worldContent.GetInventoryItemDefinition(int32(weaponId))
		itemType := def.ItemTypeDisplayName
		pageContent.WeaponDefs[weaponId] = def
		pageContent.WeaponTypeKills[itemType] += pageContent.WeaponKills[weaponId]
	}
	allWeaponsTypesJson, _ := json.Marshal(pageContent.WeaponTypeKills)
	pageContent.JsonData["allWeaponsTypes"] = string(allWeaponsTypesJson)

	// cwd, _ := os.Getwd()
	tplPath := cwd + "/web/tmpl/activities.html"
	tpl := template.Must(template.New("activities.html").Funcs(template.FuncMap{
		"isSameCharacter": func(characterId int64, entryCharacterId json.Number) bool {
			entryChar, _ := entryCharacterId.Int64()
			return characterId == entryChar
		},
		"numToStr": func(num json.Number) string {
			return num.String()
		},
	}).ParseFiles(tplPath))
	if tplErr := tpl.Execute(w, pageContent); tplErr != nil {
		fmt.Println(("exec err"))
		log.Fatal(tplErr)
	}
}

func characterHandler(w http.ResponseWriter, r *http.Request) {
	destinyCharacterId := viper.GetInt("destiny_character_id")
	destinyMembershipId := viper.GetInt("destiny_membership_id")
	components := []string{"Character"}
	content := ProfileContent{}
	content.CharacterData = requestCharacter(int32(DestinyMembershipType), int64(destinyMembershipId), int64(destinyCharacterId), components)
	tplPath := cwd + "/web/tmpl/character.html"
	tpl := template.Must(template.New("character.html").Funcs(template.FuncMap{
		"numToStr": func(num json.Number) string {
			return num.String()
		},
	}).ParseFiles(tplPath))
	if tplErr := tpl.Execute(w, content); tplErr != nil {
		fmt.Println(("exec err"))
		log.Fatal(tplErr)
	}
}

func main() {

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("failed to load config file %s", err)
	}

	if viper.GetBool("request_manifest") {
		requestDestinyManifest()
	}

	db, err := sql.Open("sqlite3", viper.GetString("manifest_path"))
	if err != nil {
		log.Fatal(err)
	}

	worldContent = app.NewSqliteDB(db)
	fmt.Println("got world content")

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/character", characterHandler)
	mux.HandleFunc("/activities", activitiesHandler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(HttpPort), mux))
}

func requestActivityHistory(membershipType int, destinyMembershipId int, characterId int, mode int, count int) app.DestinyActivityHistoryResults {
	url := fmt.Sprintf("%s/Platform/Destiny2/%d/Account/%d/Character/%d/Stats/Activities/?mode=%d&count=%d",
		ApiRootPath, membershipType, destinyMembershipId, characterId, mode, count)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("X-API-Key", viper.GetString("bungie_api_key"))
	res, reqErr := client.Do(req)
	if reqErr != nil {
		log.Fatal((reqErr))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	myactivities := app.ActivityResponse{}
	jsonErr := json.Unmarshal(body, &myactivities)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// fmt.Printf(string(body))

	return myactivities.Response
}

func requestPostGameCarnageReport(activtyId int64) app.DestinyPostGameCarnageReportData {
	url := fmt.Sprintf("%s/Platform/Destiny2/Stats/PostGameCarnageReport/%d/", ApiRootPath, activtyId)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("X-API-Key", viper.GetString("bungie_api_key"))
	res, reqErr := client.Do(req)
	if reqErr != nil {
		log.Fatal((reqErr))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	report := app.PGCRResponse{}
	jsonErr := json.Unmarshal(body, &report)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// fmt.Printf(string(body))

	return report.Response
}

func requestDestinyManifest() app.ConfigDestinyManifest {
	url := fmt.Sprintf("%s/Platform/Destiny2/Manifest/", ApiRootPath)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("X-API-Key", viper.GetString("bungie_api_key"))
	res, reqErr := client.Do(req)
	if reqErr != nil {
		log.Fatal((reqErr))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	config := app.ConfigResponse{}
	jsonErr := json.Unmarshal(body, &config)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// fmt.Printf(string(body))
	fmt.Println(config.Response.Version)
	return config.Response
}

func requestCharacter(membershipType int32, membershipId int64, characterId int64, components []string) app.DestinyCharacterResponse {
	url := fmt.Sprintf("%s/Platform/Destiny2/%d/Profile/%d/Character/%d/", ApiRootPath, membershipType, membershipId, characterId)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("components", strings.Join(components, ","))
	req.URL.RawQuery = query.Encode()

	req.Header.Set("X-API-Key", viper.GetString("bungie_api_key"))
	res, reqErr := client.Do(req)
	if reqErr != nil {
		log.Fatal((reqErr))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	result := app.CharacterResponse{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// fmt.Printf(string(body))

	return result.Response
}

func requestProfile(membershipType int32, membershipId int64, components []string) app.DestinyProfileResponse {
	url := fmt.Sprintf("%s/Platform/Destiny2/%d/Profile/%d/", ApiRootPath, membershipType, membershipId)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("components", strings.Join(components, ","))
	req.URL.RawQuery = query.Encode()

	fmt.Printf("url %s\n", req.URL)

	req.Header.Set("X-API-Key", viper.GetString("bungie_api_key"))
	res, reqErr := client.Do(req)
	if reqErr != nil {
		fmt.Println("send err")
		log.Fatal((reqErr))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Printf("readerr: %s\n", readErr)
		log.Fatal(readErr)
	}

	result := app.ProfileResponse{}
	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		fmt.Printf("json err %s\n", body)
		log.Fatal(jsonErr)
	}

	// fmt.Printf(string(body))

	return result.Response
}

func rgbContrast(red byte, green byte, blue byte, alpha byte) []int {
	
	r := int(red)
	g := int(green)
	b := int(blue)
	a := int(alpha)
	// rgb := fmt.Sprintf("rgb(%d,%d,%d,%d)", 255 - r, 255 - g, 255 - b, a)
	// return rgb
	rgb := []int{
		255 - r,
		255 - g,
		255 - b,
		a/255,
	}
	return rgb
}