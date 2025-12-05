package main

import (
    _ "embed"
    "fmt"
    "html/template"
    "log"
    "math"
    "net/http"
    "strconv"
)

//go:embed index.html
var indexHTML string

var tmpl = template.Must(template.New("index").Parse(indexHTML))

type Result struct {
    CoolingKW float64
    Model     string
    Details   string
}

func main() {
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/calculate", calculateHandler)
    fmt.Println("Калькулятор сплит-систем запущен на http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl.Execute(w, nil)
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    area := parseFloat(r.FormValue("area"))
    height := parseFloatOr(r.FormValue("height"), 2.7)
    sun := r.FormValue("sun") == "yes"
    people := parseInt(r.FormValue("people"))
    equipPower := parseFloat(r.FormValue("equip_power")) // используем напрямую
    climate := r.FormValue("climate") == "hot"

    // === РАСЧЁТ ===
    base := 100.0
    if sun {
        base += 30
    } else {
        base += 10
    }
    if climate {
        base += 20
    }

    heat := area*base +
        float64(people)*120 +
        equipPower +
        area*10 // освещение

    totalKW := math.Round((heat*1.2)/1000*10) / 10 // +20% запас, округление до 0.1 кВт

    var model string
    switch {
    case totalKW <= 2.2:
        model = "07-ка (2.0–2.2 кВт)"
    case totalKW <= 2.8:
        model = "09-ка (2.5–2.8 кВт)"
    case totalKW <= 3.8:
        model = "12-ка (3.5 кВт)"
    case totalKW <= 5.5:
        model = "18-ка (5.0–5.3 кВт)"
    case totalKW <= 7.5:
        model = "24-ка (7.0–7.1 кВт)"
    case totalKW <= 10.5:
        model = "30 / 36 (8–10 кВт)"
    default:
        model = "Кассетная или канальная (10+ кВт)"
    }

    details := fmt.Sprintf(`
Площадь: %.0f м² × %.1f м<br>
Солнечная сторона: %v<br>
Людей: %d<br>
Тепло от техники: %.0f Вт<br>
Жаркий климат: %v
    `, area, height, yesNo(sun), people, equipPower, yesNo(climate))

    result := Result{
        CoolingKW: totalKW,
        Model:     model,
        Details:   details,
    }

    w.Header().Set("Content-Type", "text/html")
    tmpl.ExecuteTemplate(w, "result", result)
}

// ─────── Вспомогательные функции ───────
func parseFloat(s string) float64 {
    f, _ := strconv.ParseFloat(s, 64)
    if f <= 0 {
        return 0
    }
    return f
}

func parseFloatOr(s string, def float64) float64 {
    if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 {
        return f
    }
    return def
}

func parseInt(s string) int {
    i, _ := strconv.Atoi(s)
    if i < 0 {
        return 0
    }
    return i
}

func yesNo(b bool) string {
    if b {
        return "Да"
    }
    return "Нет"
}