package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// ============ CORE NYA ============
func decimalToBinaryIterative(n int) string {
	if n == 0 {
		return "0"
	}

	binary := ""
	for n > 0 {
		binary = string('0'+n%2) + binary
		n = n / 2
	}
	return binary
}

func decimalToBinaryRecursive(n int) string {
	if n == 0 {
		return "0"
	}
	if n == 1 {
		return "1"
	}
	return decimalToBinaryRecursive(n/2) + string('0'+n%2)
}

// ================================================

// ============ FUNGSI DENGAN STEPS ============
func decimalToBinaryIterativeWithSteps(n int) (string, int) {
	if n == 0 {
		return "0", 1
	}

	binary := ""
	steps := 0
	temp := n
	for temp > 0 {
		steps++
		binary = string('0'+temp%2) + binary
		temp = temp / 2
	}
	return binary, steps
}

func decimalToBinaryRecursiveWithSteps(n int) (string, int) {
	steps := 0

	var recursiveHelper func(int) string
	recursiveHelper = func(x int) string {
		steps++
		if x == 0 {
			return "0"
		}
		if x == 1 {
			return "1"
		}
		return recursiveHelper(x/2) + string('0'+x%2)
	}

	return recursiveHelper(n), steps
}

type Point struct {
	N          int   `json:"n"`
	TimeIt     int64 `json:"timeIterative"`
	TimeRecur  int64 `json:"timeRecursive"`
	StepsIt    int   `json:"stepsIterative"`
	StepsRecur int   `json:"stepsRecursive"`
}

// FUNGSI UKUR WAKTU DENGAN MINIMAL 10 NS
func measureTimeIterative(n int) int64 {
	iterations := getIterations(n)

	
	for i := 0; i < 1000; i++ {
		decimalToBinaryIterative(n)
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		decimalToBinaryIterative(n)
	}
	elapsed := time.Since(start)

	if iterations > 0 {
		result := elapsed.Nanoseconds() / int64(iterations)
		// PASTIKAN MINIMAL 10 NS
		if result <= 0 {
			return 10
		}
		return result
	}
	return 10
}

func measureTimeRecursive(n int) int64 {
	if n > 1000000 || n <= 0 {
		return 10
	}

	iterations := getIterations(n)
	if iterations > 10000 {
		iterations = 10000
	}

	
	for i := 0; i < 1000 && i < iterations; i++ {
		decimalToBinaryRecursive(n)
	}

	start := time.Now()
	for i := 0; i < iterations; i++ {
		decimalToBinaryRecursive(n)
	}
	elapsed := time.Since(start)

	if iterations > 0 {
		result := elapsed.Nanoseconds() / int64(iterations)
		// PASTIKAN MINIMAL 10 NS
		if result <= 0 {
			return 10
		}
		return result
	}
	return 10
}

// FUNGSI GET ITERATIONS
func getIterations(n int) int {
	if n <= 0 {
		return 1000000
	}

	digitCount := 0
	temp := n
	for temp > 0 {
		digitCount++
		temp = temp >> 1
	}

	iterations := 10000000 / (digitCount*digitCount + 100)

	if iterations < 100 {
		iterations = 100
	}
	if iterations > 1000000 {
		iterations = 1000000
	}

	return iterations
}

// BENCHMARK FUNCTION
func benchmark(n int) Point {
	// Ukur steps dan binary
	binaryIt, stepsIt := decimalToBinaryIterativeWithSteps(n)

	var binaryRec string
	var stepsRec int

	if n > 1000000 {
		binaryRec = "N/A"
		stepsRec = 0
	} else {
		binaryRec, stepsRec = decimalToBinaryRecursiveWithSteps(n)

		// Validasi hasil sama
		if binaryIt != binaryRec {
			log.Printf("WARNING: Konversi berbeda untuk n=%d", n)
		}
	}

	// Ukur waktu
	timeIt := measureTimeIterative(n)
	timeRec := measureTimeRecursive(n)

	// PASTIKAN TIDAK ADA YANG 0 UNTUK GRAFIK
	if timeIt <= 0 {
		timeIt = 10
	}
	if timeRec <= 0 {
		timeRec = 10
	}

	_ = binaryIt
	_ = binaryRec

	return Point{n, timeIt, timeRec, stepsIt, stepsRec}
}

// API HANDLER
func apiRun(w http.ResponseWriter, r *http.Request) {
	nStr := r.URL.Query().Get("n")
	mode := r.URL.Query().Get("mode")

	n, err := strconv.Atoi(nStr)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Input harus angka"})
		return
	}

	// Validasi input
	if n > 1000000000 {
		n = 1000000000
	} else if n < 0 {
		n = 0
	}

	resp := map[string]interface{}{"n": n}

	// Single run - ITERATIF
	if mode == "iterative" || mode == "both" {
		binary, steps := decimalToBinaryIterativeWithSteps(n)
		time := measureTimeIterative(n)
		resp["iterative"] = map[string]interface{}{
			"binary": binary,
			"steps":  steps,
			"time":   time,
		}
	}

	// Single run - REKURSIF
	if mode == "recursive" || mode == "both" {
		if n > 1000000 {
			resp["recursive"] = map[string]interface{}{
				"binary":  "TOO_LARGE_FOR_RECURSION",
				"steps":   0,
				"time":    10,
				"warning": "Input > 1 juta tidak direkomendasikan untuk rekursif",
			}
		} else {
			binary, steps := decimalToBinaryRecursiveWithSteps(n)
			time := measureTimeRecursive(n)

			if time == 0 && steps > 0 {
				time = 10
			}

			resp["recursive"] = map[string]interface{}{
				"binary": binary,
				"steps":  steps,
				"time":   time,
			}
		}
	}

	// Data untuk grafik
	sizes := []int{}
	points := []Point{}

	testPoints := generateTestPoints(n)
	for _, testN := range testPoints {
		if testN <= 0 {
			continue
		}
		sizes = append(sizes, testN)
		points = append(points, benchmark(testN))
	}

	resp["chart"] = map[string]interface{}{
		"sizes":  sizes,
		"points": points,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GENERATE TEST POINTS
func generateTestPoints(userN int) []int {
	maxRange := userN
	if maxRange < 100 {
		maxRange = 100
	}
	if maxRange > 1000000 {
		maxRange = 1000000
	}

	standardPoints := []int{
		1, 2, 5, 10, 20, 50,
		100, 200, 500,
		1000, 2000, 5000,
		10000, 20000, 50000,
		100000, 200000, 500000,
		1000000,
	}

	points := []int{}

	for _, p := range standardPoints {
		if p <= maxRange {
			points = append(points, p)
		}
	}

	if userN > 0 && userN <= 1000000 {
		found := false
		for _, p := range points {
			if p == userN {
				found = true
				break
			}
		}
		if !found {
			points = append(points, userN)
			sort.Ints(points)
		}
	}

	return points
}

func main() {
	fmt.Println("=== Konversi Desimal → Biner ===")
	fmt.Println("Server aktif di localhost:8080")
	fmt.Println("Ctrl+C untuk berhenti\n")

	// Start server
	http.HandleFunc("/api/run", apiRun)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

