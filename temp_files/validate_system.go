package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BOCK-CHAIN/BockChain/tests"
)

// SystemValidator provides a simple interface to validate the complete DAO system
type SystemValidator struct {
	startTime time.Time
}

// NewSystemValidator creates a new system validator
func NewSystemValidator() *SystemValidator {
	return &SystemValidator{
		startTime: time.Now(),
	}
}

// ValidateCompleteSystem runs comprehensive system validation
func (v *SystemValidator) ValidateCompleteSystem() error {
	fmt.Println("🔍 ProjectX DAO System Validation")
	fmt.Println("=" * 50)
	fmt.Printf("Start Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("=" * 50)

	// Run the comprehensive system validation
	fmt.Println("Running comprehensive system validation tests...")

	err := tests.RunSystemValidation()
	if err != nil {
		return fmt.Errorf("system validation failed: %w", err)
	}

	duration := time.Since(v.startTime)

	fmt.Println("\n" + "="*50)
	fmt.Println("✅ SYSTEM VALIDATION COMPLETED SUCCESSFULLY")
	fmt.Printf("Total Duration: %v\n", duration)
	fmt.Println("=" * 50)

	return nil
}

func main() {
	validator := NewSystemValidator()

	if err := validator.ValidateCompleteSystem(); err != nil {
		log.Printf("❌ System validation failed: %v", err)
		os.Exit(1)
	}

	fmt.Println("🎉 System validation passed! The DAO system is fully integrated and ready.")
}
