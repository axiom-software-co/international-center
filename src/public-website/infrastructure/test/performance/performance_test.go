package performance

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/infrastructure"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/platform"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/services"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/pkg/components/public-website"
	"github.com/axiom-software-co/international-center/src/public-website/infrastructure/test/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkInfrastructureComponent_Creation(b *testing.B) {
	monitor := &shared.SharedMockResourceMonitor{}
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		b.Run("env_"+env, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				err := pulumi.RunErr(func(ctx *pulumi.Context) error {
					_, err := infrastructure.NewInfrastructureComponent(ctx, "benchmark-infra", &infrastructure.InfrastructureArgs{
						Environment: env,
					})
					return err
				}, pulumi.WithMocks("project", "stack", monitor))
				
				require.NoError(b, err)
			}
		})
	}
}

func BenchmarkPlatformComponent_Creation(b *testing.B) {
	monitor := &shared.SharedMockResourceMonitor{}
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		b.Run("env_"+env, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				err := pulumi.RunErr(func(ctx *pulumi.Context) error {
					_, err := platform.NewPlatformComponent(ctx, "benchmark-platform", &platform.PlatformArgs{
						Environment: env,
						InfrastructureOutputs: pulumi.Map{},
					})
					return err
				}, pulumi.WithMocks("project", "stack", monitor))
				
				require.NoError(b, err)
			}
		})
	}
}

func BenchmarkServicesComponent_Creation(b *testing.B) {
	monitor := &shared.SharedMockResourceMonitor{}
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		b.Run("env_"+env, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				err := pulumi.RunErr(func(ctx *pulumi.Context) error {
					_, err := services.NewServicesComponent(ctx, "benchmark-services", &services.ServicesArgs{
						Environment: env,
						InfrastructureOutputs: pulumi.Map{},
						PlatformOutputs: pulumi.Map{},
					})
					return err
				}, pulumi.WithMocks("project", "stack", monitor))
				
				require.NoError(b, err)
			}
		})
	}
}

func BenchmarkWebsiteComponent_Creation(b *testing.B) {
	monitor := &shared.SharedMockResourceMonitor{}
	environments := []string{"development", "staging", "production"}
	
	for _, env := range environments {
		b.Run("env_"+env, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				err := pulumi.RunErr(func(ctx *pulumi.Context) error {
					_, err := website.NewWebsiteComponent(ctx, "benchmark-website", &website.WebsiteArgs{
						Environment: env,
						InfrastructureOutputs: pulumi.Map{},
						PlatformOutputs: pulumi.Map{},
						ServicesOutputs: pulumi.Map{},
					})
					return err
				}, pulumi.WithMocks("project", "stack", monitor))
				
				require.NoError(b, err)
			}
		})
	}
}

func TestComponentCreation_MemoryUsage(t *testing.T) {
	// Get initial memory stats
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	monitor := &shared.SharedMockResourceMonitor{}
	
	// Create multiple components to test memory usage
	for i := 0; i < 10; i++ {
		err := pulumi.RunErr(func(ctx *pulumi.Context) error {
			// Create all component types
			_, err := infrastructure.NewInfrastructureComponent(ctx, "memory-test-infra", &infrastructure.InfrastructureArgs{
				Environment: "development",
			})
			if err != nil {
				return err
			}
			
			_, err = platform.NewPlatformComponent(ctx, "memory-test-platform", &platform.PlatformArgs{
				Environment: "development",
				InfrastructureOutputs: pulumi.Map{},
			})
			if err != nil {
				return err
			}
			
			_, err = services.NewServicesComponent(ctx, "memory-test-services", &services.ServicesArgs{
				Environment: "development",
				InfrastructureOutputs: pulumi.Map{},
				PlatformOutputs: pulumi.Map{},
			})
			if err != nil {
				return err
			}
			
			_, err = website.NewWebsiteComponent(ctx, "memory-test-website", &website.WebsiteArgs{
				Environment: "development",
				InfrastructureOutputs: pulumi.Map{},
				PlatformOutputs: pulumi.Map{},
				ServicesOutputs: pulumi.Map{},
			})
			return err
		}, pulumi.WithMocks("project", "stack", monitor))
		
		require.NoError(t, err)
	}
	
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	memoryUsedMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
	t.Logf("Memory used during component creation: %.2f MB", memoryUsedMB)
	
	// Assert memory usage is within reasonable bounds (100MB for 40 components)
	assert.Less(t, memoryUsedMB, 100.0, "Memory usage should be less than 100MB for component creation tests")
}

func TestConcurrentComponentCreation(t *testing.T) {
	const numGoroutines = 4
	const componentsPerGoroutine = 5
	
	monitor := &shared.SharedMockResourceMonitor{}
	
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*componentsPerGoroutine)
	
	start := time.Now()
	
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for i := 0; i < componentsPerGoroutine; i++ {
				err := pulumi.RunErr(func(ctx *pulumi.Context) error {
					_, err := infrastructure.NewInfrastructureComponent(ctx, "concurrent-test-infra", &infrastructure.InfrastructureArgs{
						Environment: "development",
					})
					return err
				}, pulumi.WithMocks("project", "stack", monitor))
				
				if err != nil {
					errChan <- err
					return
				}
			}
		}(g)
	}
	
	wg.Wait()
	close(errChan)
	
	// Check for any errors
	for err := range errChan {
		require.NoError(t, err)
	}
	
	duration := time.Since(start)
	t.Logf("Concurrent component creation took: %v", duration)
	
	// Assert that concurrent creation completes within reasonable time (10 seconds)
	assert.Less(t, duration.Seconds(), 10.0, "Concurrent component creation should complete within 10 seconds")
}

func TestComponentCreation_ExecutionTime(t *testing.T) {
	monitor := &shared.SharedMockResourceMonitor{}
	config := shared.GetDefaultPerformanceConfig()
	
	testCases := []struct {
		name        string
		createFunc  func(*pulumi.Context) error
		environment string
	}{
		{
			name:        "Infrastructure_Development",
			environment: "development",
			createFunc: func(ctx *pulumi.Context) error {
				_, err := infrastructure.NewInfrastructureComponent(ctx, "perf-test-infra", &infrastructure.InfrastructureArgs{
					Environment: "development",
				})
				return err
			},
		},
		{
			name:        "Platform_Production",
			environment: "production",
			createFunc: func(ctx *pulumi.Context) error {
				_, err := platform.NewPlatformComponent(ctx, "perf-test-platform", &platform.PlatformArgs{
					Environment: "production",
					InfrastructureOutputs: pulumi.Map{},
				})
				return err
			},
		},
		{
			name:        "Services_Staging",
			environment: "staging",
			createFunc: func(ctx *pulumi.Context) error {
				_, err := services.NewServicesComponent(ctx, "perf-test-services", &services.ServicesArgs{
					Environment: "staging",
					InfrastructureOutputs: pulumi.Map{},
					PlatformOutputs: pulumi.Map{},
				})
				return err
			},
		},
		{
			name:        "Website_Production",
			environment: "production",
			createFunc: func(ctx *pulumi.Context) error {
				_, err := website.NewWebsiteComponent(ctx, "perf-test-website", &website.WebsiteArgs{
					Environment: "production",
					InfrastructureOutputs: pulumi.Map{},
					PlatformOutputs: pulumi.Map{},
					ServicesOutputs: pulumi.Map{},
				})
				return err
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			
			err := pulumi.RunErr(tc.createFunc, pulumi.WithMocks("project", "stack", monitor))
			
			duration := time.Since(start)
			
			require.NoError(t, err)
			
			t.Logf("Component creation for %s took: %v", tc.name, duration)
			
			maxDuration := time.Duration(config.MaxExecutionTimeMs) * time.Millisecond
			assert.Less(t, duration, maxDuration, "Component creation should complete within %v", maxDuration)
		})
	}
}