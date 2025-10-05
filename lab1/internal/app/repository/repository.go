package repository

import (
	"fmt"
	"strings"
)

type Repository struct{}

type UtilityService struct {
	ID          int
	Title       string
	Description string
	ImageURL    string
	Unit        string
	Tariff      float32
}

type SingleUtilityApplication struct {
	Service  UtilityService
	Quantity float32
	Total    float32
}

type UtilityApplication struct {
	Services  []SingleUtilityApplication
	TotalCost float32
}

var utilityServices = []UtilityService{
	{
		ID:          1,
		Title:       "Электроэнергия (кВт/ч)",
		Description: "Обеспечение жилого помещения электроэнергией. Платёж за потреблённую электроэнергию по счётчику (день/ночь).",
		ImageURL:    "http://localhost:9000/kartinki/energy.jpg",
		Unit:        "кВт/ч",
		Tariff:      3.6,
	},
	{
		ID:          2,
		Title:       "Отопление (Гкал)",
		Description: "Обеспечение теплом жилого помещения в холодное время года. Тепло поступает от централизованной системы отопления (котельная, ТЭЦ) или от индивидуального отопительного оборудования (газовый котел, электрический котел). Расчёт происходит по доле теплопотребления дома (общедомовой счётчик).",
		ImageURL:    "http://localhost:9000/kartinki/heat.jpg",
		Unit:        "Гкал",
		Tariff:      1252.31,
	},
	{
		ID:          3,
		Title:       "Холодное водоснабжение (м³)",
		Description: "Подача холодной воды в жилое помещение. Начисление за объём холодной воды по счётчику.",
		ImageURL:    "http://localhost:9000/kartinki/cold_water.jpg",
		Unit:        "м³",
		Tariff:      16.58,
	},
	{
		ID:          4,
		Title:       "Горячее водоснабжение (м³)",
		Description: "Подача горячей воды в жилое помещение. Начисление за объём горячей воды по счётчику.",
		ImageURL:    "http://localhost:9000/kartinki/hot water.jpg",
		Unit:        "м³",
		Tariff:      97.97,
	},
	{
		ID:          5,
		Title:       "Газ (кг)",
		Description: "Обеспечение жилого помещения газом. Потребление газа по счётчику.",
		ImageURL:    "http://localhost:9000/kartinki/gas.jpg",
		Unit:        "кг",
		Tariff:      26.68,
	},
	{
		ID:          6,
		Title:       "Капитальный ремонт (м²)",
		Description: "Финансирование работ по капитальному ремонту общего имущества многоквартирного дома (МКД). Включает в себя ремонт крыши, фасада, фундамента, лифтов, инженерных систем и т.д. Начисление по площади.",
		ImageURL:    "http://localhost:9000/kartinki/kap_rem.png",
		Unit:        "м²",
		Tariff:      4.5,
	},
}

var Utility__Application = map[int]UtilityApplication{
	1: {
		TotalCost: 0,
		Services: []SingleUtilityApplication{
			{Service: utilityServices[0], Quantity: 300, Total: 300 * utilityServices[0].Tariff},
			{Service: utilityServices[1], Quantity: 1, Total: 1 * utilityServices[1].Tariff},
			{Service: utilityServices[5], Quantity: 50, Total: 50 * utilityServices[5].Tariff},
		},
	},
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

func (r *Repository) GetUtility__Application(id int) (UtilityApplication, error) {
	c, ok := Utility__Application[id]
	if !ok {
		return UtilityApplication{}, fmt.Errorf("корзина не найдена")
	}
	var total float32
	for _, s := range c.Services {
		total += s.Total
	}
	c.TotalCost = total
	return c, nil
}

func (r *Repository) GetUtilityServices() ([]UtilityService, error) {
	if len(utilityServices) == 0 {
		return nil, fmt.Errorf("список услуг пуст")
	}
	return utilityServices, nil
}

func (r *Repository) GetUtilityServiceByID(id int) (UtilityService, error) {
	for _, s := range utilityServices {
		if s.ID == id {
			return s, nil
		}
	}
	return UtilityService{}, fmt.Errorf("услуга не найдена")
}

func (r *Repository) SearchUtilityServices(title string) ([]UtilityService, error) {
	services, err := r.GetUtilityServices()
	if err != nil {
		return nil, err
	}
	title = strings.ToLower(strings.TrimSpace(title))
	if title == "" {
		return services, nil
	}
	var result []UtilityService
	for _, s := range services {
		if strings.Contains(strings.ToLower(s.Title), title) {
			result = append(result, s)
		}
	}
	return result, nil
}
