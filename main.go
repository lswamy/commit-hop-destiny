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

var worldContent *app.SQLiteDB

type PageContent struct {
	Name            string
	MainCharacterId int64
	Activities      []ActivityData
	WeaponKills     map[int64]float32
	WeaponDefs      map[int64]app.InventoryItemDefinition
	CharacterStats  struct {
		WeaponKills map[int64]float32
	}
	JsonData map[string]string
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
	characterActivities := requestActivityHistory(DestinyMembershipType, viper.GetInt("destiny_membership_id"), destinyCharacterId, ActivityModeTypePvp)

	pageContent := PageContent{}
	pageContent.MainCharacterId = int64(destinyCharacterId)
	pageContent.Activities = make([]ActivityData, 0)
	pageContent.WeaponKills = make(map[int64]float32)
	pageContent.CharacterStats.WeaponKills = make(map[int64]float32)
	pageContent.JsonData = make(map[string]string)

	s, _ := json.MarshalIndent(pageContent, "", "\t")
	fmt.Print(string(s))

	for a := range characterActivities.Activities {
		ad := ActivityData{}
		activity := characterActivities.Activities[a]
		activityHash := int32(activity.ActivityDetails.DirectorActivityHash)
		referenceId := int32(activity.ActivityDetails.ReferenceId)
		instanceId, _ := activity.ActivityDetails.InstanceId.Int64()

		activityDef := worldContent.GetActivityDefinition(activityHash)
		s, _ := json.MarshalIndent(activityDef, "", "\t")
		fmt.Print(string(s))

		if activityHash != referenceId {
			ad.ReferenceDefinition = worldContent.GetActivityDefinition(referenceId)
		}

		ad.InstanceId = instanceId
		ad.Definition = activityDef
		ad.Period = activity.Period

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
		pageContent.WeaponDefs[weaponId] = def
	}

	cwd, _ := os.Getwd()
	tplPath := cwd + "/web/tmpl/index.html"
	tpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
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
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(HttpPort), mux))
}

func requestActivityHistory(membershipType int, destinyMembershipId int, characterId int, mode int) app.DestinyActivityHistoryResults {
	url := fmt.Sprintf("%s/Platform/Destiny2/%d/Account/%d/Character/%d/Stats/Activities/?mode=%d",
		ApiRootPath, membershipType, destinyMembershipId, characterId, mode)

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
