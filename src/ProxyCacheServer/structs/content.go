package structs

import (
	"database/sql"

	"github.com/asaskevich/govalidator"
)

// Структура таблицы Content
type Content struct {
	Id                    int            `db:"id" valid:"required"`                                // Первичный ключ
	StationNameRu         sql.NullString `db:"stationNameRu" valid:"-"`                            // Название станции (на русском языке)
	StationNameEn         sql.NullString `db:"stationNameEn" valid:"-"`                            // Название станции (на английском языке)
	StationCode           string         `db:"stationCode" valid:"required,runelength(1|7)"`       // Код станции
	StationCityId         sql.NullInt64  `db:"stationCityId" valid:"-"`                            // Первичный ключ города, к которому привязана станция
	StationRegionId       sql.NullInt64  `db:"stationRegionId" valid:"-"`                          // Первичный ключ региона, к которому привязана станция
	StationCountryId      sql.NullInt64  `db:"stationCountryId" valid:"-"`                         // Первичный ключ страны, к которому привязана станция
	StationIsAutocomplete string         `db:"stationIsAutocomplete" valid:"required,length(1|1)"` // Учавствует ли запись в авткомплите
	StationPriority       sql.NullInt64  `db:"stationPriority" valid:"-"`                          // Station Priority
	CityName              sql.NullString `db:"cityName" valid:"-"`                                 // Название населённого пункта
	CityTypeId            sql.NullInt64  `db:"cityTypeId" valid:"-"`                               // Первичный ключ типа населённого пункта, к которому привязана станция
	CityTypeName          sql.NullString `db:"cityTypeName" valid:"-"`                             // Название типа города
	RegionName            sql.NullString `db:"regionName" valid:"-"`                               // Название региона
	CountryName           sql.NullString `db:"countryName" valid:"-"`                              // Название страны
	StationAliasName      sql.NullString `db:"stationAliasName" valid:"-"`                         // Название алиаса станции
}

// Конструктор структуры Content
func NewContent(id int, stationCode string, stationIsAutocomplete string) *Content {
	// Определяем значения по умолчанию
	return &Content{
		Id:                    id,
		StationCode:           stationCode,
		StationIsAutocomplete: stationIsAutocomplete,
	}
}

// Валидатор структуры Content
func (this *Content) Validate() (bool, error) {
	return govalidator.ValidateStruct(this)
}
