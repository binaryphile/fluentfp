package main

import "github.com/binaryphile/fluentfp/slice"

// Report generator module - processes employee data for quarterly reports.
// All operations are data transformations (filter/map/fold).

type Employee struct {
	ID, Name, Email, Department, Manager string
	Salary, Bonus                        float64
	YearsOfService                       int
	Active, Remote, FullTime             bool
	PerformanceScore                     float64
}

type DepartmentSummary struct {
	Name         string
	HeadCount    int
	TotalSalary  float64
	AvgScore     float64
	RemoteCount  int
}

// Predicates
func (e Employee) IsActive() bool        { return e.Active }
func (e Employee) IsRemote() bool        { return e.Remote }
func (e Employee) IsFullTime() bool      { return e.FullTime }
func (e Employee) IsHighPerformer() bool { return e.PerformanceScore >= 4.0 }
func (e Employee) IsLowPerformer() bool  { return e.PerformanceScore < 2.5 }
func (e Employee) IsSenior() bool        { return e.YearsOfService >= 5 }
func (e Employee) IsNewHire() bool       { return e.YearsOfService < 1 }
func (e Employee) IsBonusEligible() bool { return e.Active && e.PerformanceScore >= 3.0 }

// Getters
func (e Employee) GetID() string               { return e.ID }
func (e Employee) GetName() string             { return e.Name }
func (e Employee) GetEmail() string            { return e.Email }
func (e Employee) GetDepartment() string       { return e.Department }
func (e Employee) GetManager() string          { return e.Manager }
func (e Employee) GetSalary() float64          { return e.Salary }
func (e Employee) GetBonus() float64           { return e.Bonus }
func (e Employee) GetCompensation() float64    { return e.Salary + e.Bonus }
func (e Employee) GetYearsOfService() int      { return e.YearsOfService }
func (e Employee) GetPerformanceScore() float64 { return e.PerformanceScore }

var sumFloat = func(a, b float64) float64 { return a + b }
var sumInt = func(a, b int) int { return a + b }

// === Filtering Functions ===

func getActiveEmployees(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsActive)
}

func getRemoteWorkers(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsRemote)
}

func getFullTimeEmployees(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsFullTime)
}

func getHighPerformers(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsHighPerformer)
}

func getLowPerformers(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsLowPerformer)
}

func getSeniorEmployees(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsSenior)
}

func getNewHires(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsNewHire)
}

func getBonusEligible(employees []Employee) []Employee {
	return slice.From(employees).KeepIf(Employee.IsBonusEligible)
}

func getByDepartment(employees []Employee, dept string) []Employee {
	return slice.From(employees).KeepIf(func(e Employee) bool { return e.Department == dept })
}

func getByManager(employees []Employee, manager string) []Employee {
	return slice.From(employees).KeepIf(func(e Employee) bool { return e.Manager == manager })
}

// === Extraction Functions ===

func getEmployeeIDs(employees []Employee) []string {
	return slice.From(employees).ToString(Employee.GetID)
}

func getEmployeeNames(employees []Employee) []string {
	return slice.From(employees).ToString(Employee.GetName)
}

func getEmployeeEmails(employees []Employee) []string {
	return slice.From(employees).ToString(Employee.GetEmail)
}

func getManagerNames(employees []Employee) []string {
	return slice.From(employees).ToString(Employee.GetManager)
}

func getDepartmentNames(employees []Employee) []string {
	return slice.From(employees).ToString(Employee.GetDepartment)
}

// === Counting Functions ===

func countActive(employees []Employee) int {
	return slice.From(employees).KeepIf(Employee.IsActive).Len()
}

func countRemote(employees []Employee) int {
	return slice.From(employees).KeepIf(Employee.IsRemote).Len()
}

func countFullTime(employees []Employee) int {
	return slice.From(employees).KeepIf(Employee.IsFullTime).Len()
}

func countHighPerformers(employees []Employee) int {
	return slice.From(employees).KeepIf(Employee.IsHighPerformer).Len()
}

func countBonusEligible(employees []Employee) int {
	return slice.From(employees).KeepIf(Employee.IsBonusEligible).Len()
}

// === Aggregation Functions ===

func totalSalary(employees []Employee) float64 {
	return slice.Fold(slice.From(employees).ToFloat64(Employee.GetSalary), 0.0, sumFloat)
}

func totalBonus(employees []Employee) float64 {
	return slice.Fold(slice.From(employees).ToFloat64(Employee.GetBonus), 0.0, sumFloat)
}

func totalCompensation(employees []Employee) float64 {
	return slice.Fold(slice.From(employees).ToFloat64(Employee.GetCompensation), 0.0, sumFloat)
}

func averageSalary(employees []Employee) float64 {
	sum := slice.Fold(slice.From(employees).ToFloat64(Employee.GetSalary), 0.0, sumFloat)
	return sum / float64(len(employees))
}

func averagePerformance(employees []Employee) float64 {
	sum := slice.Fold(slice.From(employees).ToFloat64(Employee.GetPerformanceScore), 0.0, sumFloat)
	return sum / float64(len(employees))
}

func averageYearsOfService(employees []Employee) float64 {
	sum := slice.Fold(slice.From(employees).ToInt(Employee.GetYearsOfService), 0, sumInt)
	return float64(sum) / float64(len(employees))
}

func totalYearsOfService(employees []Employee) int {
	return slice.Fold(slice.From(employees).ToInt(Employee.GetYearsOfService), 0, sumInt)
}

// === Combined Operations ===

func getHighPerformerEmails(employees []Employee) []string {
	return slice.From(employees).KeepIf(Employee.IsHighPerformer).ToString(Employee.GetEmail)
}

func getRemoteWorkerNames(employees []Employee) []string {
	return slice.From(employees).KeepIf(Employee.IsRemote).ToString(Employee.GetName)
}

func getBonusEligibleIDs(employees []Employee) []string {
	return slice.From(employees).KeepIf(Employee.IsBonusEligible).ToString(Employee.GetID)
}

func totalHighPerformerSalary(employees []Employee) float64 {
	return slice.Fold(slice.From(employees).KeepIf(Employee.IsHighPerformer).ToFloat64(Employee.GetSalary), 0.0, sumFloat)
}

func averageRemoteSalary(employees []Employee) float64 {
	remotes := slice.From(employees).KeepIf(Employee.IsRemote)
	sum := slice.Fold(remotes.ToFloat64(Employee.GetSalary), 0.0, sumFloat)
	return sum / float64(remotes.Len())
}
