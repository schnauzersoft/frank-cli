package kubernetes

import (
	"log/slog"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNeedsUpdate(t *testing.T) {
	deployer := &Deployer{
		logger: slog.Default(),
	}

	tests := []struct {
		name     string
		existing *unstructured.Unstructured
		desired  *unstructured.Unstructured
		expected bool
	}{
		{
			name: "identical resources should not need update",
			existing: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
									},
								},
							},
						},
					},
				},
			},
			desired: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "different replicas should need update",
			existing: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
									},
								},
							},
						},
					},
				},
			},
			desired: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(5),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "different image should need update",
			existing: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
									},
								},
							},
						},
					},
				},
			},
			desired: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.21",
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "different port protocol should need update",
			existing: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
										"ports": []interface{}{
											map[string]interface{}{
												"containerPort": int64(80),
												"protocol":      "TCP",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			desired: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
										"ports": []interface{}{
											map[string]interface{}{
												"containerPort": int64(80),
												"protocol":      "UDP",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "missing protocol should default to TCP and match",
			existing: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
										"ports": []interface{}{
											map[string]interface{}{
												"containerPort": int64(80),
												"protocol":      "TCP",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			desired: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name": "test-deployment",
					},
					"spec": map[string]interface{}{
						"replicas": int64(3),
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":  "nginx",
										"image": "nginx:1.20",
										"ports": []interface{}{
											map[string]interface{}{
												"containerPort": int64(80),
												// protocol missing, should default to TCP
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deployer.needsUpdate(tt.existing, tt.desired)
			if result != tt.expected {
				t.Errorf("needsUpdate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComparePorts(t *testing.T) {
	deployer := &Deployer{
		logger: slog.Default(),
	}

	tests := []struct {
		name     string
		existing []interface{}
		desired  []interface{}
		expected bool
	}{
		{
			name: "identical ports should match",
			existing: []interface{}{
				map[string]interface{}{
					"containerPort": int64(80),
					"protocol":      "TCP",
				},
			},
			desired: []interface{}{
				map[string]interface{}{
					"containerPort": int64(80),
					"protocol":      "TCP",
				},
			},
			expected: true,
		},
		{
			name: "different ports should not match",
			existing: []interface{}{
				map[string]interface{}{
					"containerPort": int64(80),
					"protocol":      "TCP",
				},
			},
			desired: []interface{}{
				map[string]interface{}{
					"containerPort": int64(8080),
					"protocol":      "TCP",
				},
			},
			expected: false,
		},
		{
			name: "missing protocol should default to TCP",
			existing: []interface{}{
				map[string]interface{}{
					"containerPort": int64(80),
					"protocol":      "TCP",
				},
			},
			desired: []interface{}{
				map[string]interface{}{
					"containerPort": int64(80),
					// protocol missing, should default to TCP
				},
			},
			expected: true,
		},
		{
			name: "different protocols should not match",
			existing: []interface{}{
				map[string]interface{}{
					"containerPort": int64(80),
					"protocol":      "TCP",
				},
			},
			desired: []interface{}{
				map[string]interface{}{
					"containerPort": int64(80),
					"protocol":      "UDP",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deployer.comparePorts(tt.existing, tt.desired)
			if result != tt.expected {
				t.Errorf("comparePorts() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMapsEqual(t *testing.T) {
	deployer := &Deployer{
		logger: slog.Default(),
	}

	tests := []struct {
		name     string
		map1     map[string]interface{}
		map2     map[string]interface{}
		expected bool
	}{
		{
			name: "identical maps should be equal",
			map1: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			map2: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			expected: true,
		},
		{
			name: "different values should not be equal",
			map1: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			map2: map[string]interface{}{
				"key1": "value1",
				"key2": 43,
			},
			expected: false,
		},
		{
			name: "different keys should not be equal",
			map1: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			map2: map[string]interface{}{
				"key1": "value1",
				"key3": 42,
			},
			expected: false,
		},
		{
			name: "nested maps should be compared",
			map1: map[string]interface{}{
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
			map2: map[string]interface{}{
				"nested": map[string]interface{}{
					"inner": "value",
				},
			},
			expected: true,
		},
		{
			name: "different nested maps should not be equal",
			map1: map[string]interface{}{
				"nested": map[string]interface{}{
					"inner": "value1",
				},
			},
			map2: map[string]interface{}{
				"nested": map[string]interface{}{
					"inner": "value2",
				},
			},
			expected: false,
		},
		{
			name: "slices should be compared",
			map1: map[string]interface{}{
				"slice": []interface{}{"a", "b", "c"},
			},
			map2: map[string]interface{}{
				"slice": []interface{}{"a", "b", "c"},
			},
			expected: true,
		},
		{
			name: "different slices should not be equal",
			map1: map[string]interface{}{
				"slice": []interface{}{"a", "b", "c"},
			},
			map2: map[string]interface{}{
				"slice": []interface{}{"a", "b", "d"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deployer.mapsEqual(tt.map1, tt.map2)
			if result != tt.expected {
				t.Errorf("mapsEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}
