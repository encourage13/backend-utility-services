package repository

type Service struct {
	ID          int
	Title       string
	Description string
	ImageURL    string
	Unit        string
	Tariff      float32
	Consumption float32
	Total       float32
}

type Repository struct {
	services []Service
	cart     []int
}

func NewRepository() (*Repository, error) {
	return &Repository{
		services: []Service{
			{ID: 1, Title: "Электроэнергия (кВт/ч)", Description: "Обеспечение жилого помещения электроэнергией. Платёж за потреблённую электроэнергию по счётчику (день/ночь).", ImageURL: "http://localhost:9000/kartinki/energy.jpg", Unit: "кВт/ч", Tariff: 3.6, Consumption: 300},
			{ID: 2, Title: "Отопление (Гкал)", Description: "Обеспечение теплом жилого помещения в холодное время года. Тепло поступает от централизованной системы отопления (котельная, ТЭЦ) или от индивидуального отопительного оборудования (газовый котел, электрический котел). Расчёт происходит по доле теплопотребления дома (общедомовой счётчик)", ImageURL: "http://localhost:9000/kartinki/heat.jpg", Unit: "Гкал", Tariff: 1252.31, Consumption: 1},
			{ID: 3, Title: "Холодное водоснабжение (м^3)", Description: "Подача холодной воды в жилое помещение. Начисление за объём холодной воды по счётчику.", ImageURL: "http://localhost:9000/kartinki/cold_water.jpg", Unit: "м^3", Tariff: 16.58, Consumption: 8},
			{ID: 4, Title: "Горячее водоснабжение (м^3)", Description: "Подача горячей воды в жилое помещение. Начисление за объём горячей воды по счётчику.", ImageURL: "http://localhost:9000/kartinki/hot water.jpg", Unit: "м^3", Tariff: 97.97, Consumption: 4},
			{ID: 5, Title: "Газ (кг)", Description: "Обеспечение жилого помещения газом. Потребление газа по счётчику.", ImageURL: "http://localhost:9000/kartinki/gas.jpg", Unit: "кг", Tariff: 26.68, Consumption: 20.88},
			{ID: 6, Title: "Капитальный ремонт (м^2)", Description: "Финансирование работ по капитальному ремонту общего имущества многоквартирного дома (МКД). Включает в себя ремонт крыши, фасада, фундамента, лифтов, инженерных систем и т.д. Начисление по площади.", ImageURL: "http://localhost:9000/kartinki/kap_rem.png", Unit: "м^2", Tariff: 4.5, Consumption: 50},
		},

		cart: []int{2, 4, 5, 6},
	}, nil
}

func (r *Repository) ListServices() []Service {
	return r.services
}

func (r *Repository) GetServiceByID(id int) (Service, bool) {
	for _, s := range r.services {
		if s.ID == id {
			return s, true
		}
	}
	return Service{}, false
}

func (r *Repository) CartCount() int {

	return len(r.cart)
}

func (r *Repository) ListCartServices() []Service {
	var result []Service
	for _, id := range r.cart {
		if s, ok := r.GetServiceByID(id); ok {
			s.Total = s.Consumption * s.Tariff
			result = append(result, s)
		}
	}
	return result
}
