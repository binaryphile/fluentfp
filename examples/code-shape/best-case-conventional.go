package main

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

// === Filtering Functions ===

func getActiveEmployees(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.Active {
			result = append(result, e)
		}
	}
	return result
}

func getRemoteWorkers(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.Remote {
			result = append(result, e)
		}
	}
	return result
}

func getFullTimeEmployees(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.FullTime {
			result = append(result, e)
		}
	}
	return result
}

func getHighPerformers(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.PerformanceScore >= 4.0 {
			result = append(result, e)
		}
	}
	return result
}

func getLowPerformers(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.PerformanceScore < 2.5 {
			result = append(result, e)
		}
	}
	return result
}

func getSeniorEmployees(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.YearsOfService >= 5 {
			result = append(result, e)
		}
	}
	return result
}

func getNewHires(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.YearsOfService < 1 {
			result = append(result, e)
		}
	}
	return result
}

func getBonusEligible(employees []Employee) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.Active && e.PerformanceScore >= 3.0 {
			result = append(result, e)
		}
	}
	return result
}

func getByDepartment(employees []Employee, dept string) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.Department == dept {
			result = append(result, e)
		}
	}
	return result
}

func getByManager(employees []Employee, manager string) []Employee {
	var result []Employee
	for _, e := range employees {
		if e.Manager == manager {
			result = append(result, e)
		}
	}
	return result
}

// === Extraction Functions ===

func getEmployeeIDs(employees []Employee) []string {
	var ids []string
	for _, e := range employees {
		ids = append(ids, e.ID)
	}
	return ids
}

func getEmployeeNames(employees []Employee) []string {
	var names []string
	for _, e := range employees {
		names = append(names, e.Name)
	}
	return names
}

func getEmployeeEmails(employees []Employee) []string {
	var emails []string
	for _, e := range employees {
		emails = append(emails, e.Email)
	}
	return emails
}

func getManagerNames(employees []Employee) []string {
	var managers []string
	for _, e := range employees {
		managers = append(managers, e.Manager)
	}
	return managers
}

func getDepartmentNames(employees []Employee) []string {
	var depts []string
	for _, e := range employees {
		depts = append(depts, e.Department)
	}
	return depts
}

// === Counting Functions ===

func countActive(employees []Employee) int {
	count := 0
	for _, e := range employees {
		if e.Active {
			count++
		}
	}
	return count
}

func countRemote(employees []Employee) int {
	count := 0
	for _, e := range employees {
		if e.Remote {
			count++
		}
	}
	return count
}

func countFullTime(employees []Employee) int {
	count := 0
	for _, e := range employees {
		if e.FullTime {
			count++
		}
	}
	return count
}

func countHighPerformers(employees []Employee) int {
	count := 0
	for _, e := range employees {
		if e.PerformanceScore >= 4.0 {
			count++
		}
	}
	return count
}

func countBonusEligible(employees []Employee) int {
	count := 0
	for _, e := range employees {
		if e.Active && e.PerformanceScore >= 3.0 {
			count++
		}
	}
	return count
}

// === Aggregation Functions ===

func totalSalary(employees []Employee) float64 {
	sum := 0.0
	for _, e := range employees {
		sum += e.Salary
	}
	return sum
}

func totalBonus(employees []Employee) float64 {
	sum := 0.0
	for _, e := range employees {
		sum += e.Bonus
	}
	return sum
}

func totalCompensation(employees []Employee) float64 {
	sum := 0.0
	for _, e := range employees {
		sum += e.Salary + e.Bonus
	}
	return sum
}

func averageSalary(employees []Employee) float64 {
	sum := 0.0
	for _, e := range employees {
		sum += e.Salary
	}
	return sum / float64(len(employees))
}

func averagePerformance(employees []Employee) float64 {
	sum := 0.0
	for _, e := range employees {
		sum += e.PerformanceScore
	}
	return sum / float64(len(employees))
}

func averageYearsOfService(employees []Employee) float64 {
	sum := 0
	for _, e := range employees {
		sum += e.YearsOfService
	}
	return float64(sum) / float64(len(employees))
}

func totalYearsOfService(employees []Employee) int {
	sum := 0
	for _, e := range employees {
		sum += e.YearsOfService
	}
	return sum
}

// === Combined Operations ===

func getHighPerformerEmails(employees []Employee) []string {
	var emails []string
	for _, e := range employees {
		if e.PerformanceScore >= 4.0 {
			emails = append(emails, e.Email)
		}
	}
	return emails
}

func getRemoteWorkerNames(employees []Employee) []string {
	var names []string
	for _, e := range employees {
		if e.Remote {
			names = append(names, e.Name)
		}
	}
	return names
}

func getBonusEligibleIDs(employees []Employee) []string {
	var ids []string
	for _, e := range employees {
		if e.Active && e.PerformanceScore >= 3.0 {
			ids = append(ids, e.ID)
		}
	}
	return ids
}

func totalHighPerformerSalary(employees []Employee) float64 {
	sum := 0.0
	for _, e := range employees {
		if e.PerformanceScore >= 4.0 {
			sum += e.Salary
		}
	}
	return sum
}

func averageRemoteSalary(employees []Employee) float64 {
	sum := 0.0
	count := 0
	for _, e := range employees {
		if e.Remote {
			sum += e.Salary
			count++
		}
	}
	return sum / float64(count)
}
