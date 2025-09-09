package repository

type Service struct {
	ID          int
	Title       string
	Description string
	ImageURL    string
}

type Repository struct {
	services []Service
}

func NewRepository() (*Repository, error) {
	return &Repository{
		services: []Service{
			{ID: 1, Title: "Электроэнергия", Description: "Платёж за потреблённую электроэнергию по счётчику (день/ночь) или по нормативу при отсутствии счётчика. Возможен учёт доли ОДН.", ImageURL: "http://localhost:9000/kartinki/energy.jpg"},
			{ID: 2, Title: "Отопление", Description: "Расчёт по доле теплопотребления дома (общедомовой счётчик) или по нормативу", ImageURL: "http://localhost:9000/kartinki/heat.jpg"},
			{ID: 3, Title: "Холодное водоснабжение", Description: "Начисление за объём холодной воды по счётчику либо по нормативу на человека", ImageURL: "http://localhost:9000/kartinki/cold_water.jpg"},
			{ID: 4, Title: "Горячее водоснабжение", Description: "Начисление за объём горячей воды по счётчику либо по нормативу на человека", ImageURL: "http://localhost:9000/kartinki/hot water.jpg"},
			{ID: 5, Title: "Газ", Description: "Потребление газа по счётчику или нормативу.", ImageURL: "http://localhost:9000/kartinki/gas.jpg"},
			{ID: 6, Title: "Вывоз мусора", Description: "Начисление по числу зарегистрированных жильцов (иногда по площади).", ImageURL: "http://localhost:9000/kartinki/bin.jpg"},
		},
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
	return len(r.services)
}
