/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package create

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestCreateService(t *testing.T) {
	tests := []struct {
		options  *ServiceClusterIPOpts
		expected *v1.Service
		expectErr   bool
	}{
		{
			options: &ServiceClusterIPOpts{
				ServiceCommon{
	            Name:        "clusterip-ok",
				TCP:         []string{"456", "321:908"},
				ClusterIP:   "",
				Type: v1.ServiceTypeClusterIP,
				},
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "clusterip-ok",
					Labels: map[string]string{"app": "clusterip-ok"},
				},
				Spec: v1.ServiceSpec{Type: "ClusterIP",
					Ports: []v1.ServicePort{{Name: "456", Protocol: "TCP", Port: 456, TargetPort: intstr.IntOrString{Type: 0, IntVal: 456, StrVal: ""}, NodePort: 0},
						{Name: "321-908", Protocol: "TCP", Port: 321, TargetPort: intstr.IntOrString{Type: 0, IntVal: 908, StrVal: ""}, NodePort: 0}},
					Selector:  map[string]string{"app": "clusterip-ok"},
					ClusterIP: "", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},
		{
			options: &ServiceClusterIPOpts{
				ServiceCommon{
			  Name:        "clusterip-missing",
			  Type: v1.ServiceTypeClusterIP,
				},
			},
			expectErr:   true,
		},
		{
			options: &ServiceClusterIPOpts{
				
			ServiceCommon{
			  Name:        "clusterip-none-wrong-type",
			  TCP:         []string{},
			  ClusterIP:   "None",
			  Type: v1.ServiceTypeNodePort,
			},		
			},
			expectErr:   true,
		},		
		{
			options: &ServiceClusterIPOpts{
	            ServiceCommon{
	            Name:        "clusterip-none-ok",
				TCP:         []string{},
				ClusterIP:   "None",
				Type: v1.ServiceTypeClusterIP,
	            },
			},
			
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "clusterip-none-ok",
					Labels: map[string]string{"app": "clusterip-none-ok"},
				},
				Spec: v1.ServiceSpec{Type: "ClusterIP",
					Ports:     []v1.ServicePort{},
					Selector:  map[string]string{"app": "clusterip-none-ok"},
					ClusterIP: "None", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},
		{
			options: &ServiceClusterIPOpts{
	            ServiceCommon{
	            Name:        "clusterip-none-and-port-mapping",
				TCP:         []string{"456:9898"},
				ClusterIP:   "None",
				Type: v1.ServiceTypeClusterIP,
	            },
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "clusterip-none-and-port-mapping",
					Labels: map[string]string{"app": "clusterip-none-and-port-mapping"},
				},
				Spec: v1.ServiceSpec{Type: "ClusterIP",
					Ports:     []v1.ServicePort{{Name: "456-9898", Protocol: "TCP", Port: 456, TargetPort: intstr.IntOrString{Type: 0, IntVal: 9898, StrVal: ""}, NodePort: 0}},
					Selector:  map[string]string{"app": "clusterip-none-and-port-mapping"},
					ClusterIP: "None", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},		
//		{
//			options: &ServiceClusterIPOpts{
//				
//				ServiceCommon{
//			      Name:         "invalid-port",
//				  TCP:         []string{"65536:1"},
//				  ClusterIP:   "None",
//			      Type: v1.ServiceTypeClusterIP,
//				},
//			},
//			expectErr:   true,
//		},
//		{
//			options: &ServiceClusterIPOpts{
//				
//			ServiceCommon{
//			  Name:        "invalid-port-mapping",
//			  TCP:         []string{"8080:-abc"},
//			  ClusterIP:   "None",
//			  Type: v1.ServiceTypeClusterIP,
//			},		
//			},
//			expectErr:   true,
//		},				
	}

	for _, tc := range tests {
		service, err := tc.options.createService()
		if err != nil {
			t.Errorf("unexpected error:\n%#v\n", err)
			return
		}
		if tc.expectErr && err == nil {
		   continue
	    }
		if !apiequality.Semantic.DeepEqual(service, tc.expected) {
			t.Errorf("expected:\n%#v\ngot:\n%#v", tc.expected, service)
		}
	}
}

func TestCreateServiceNodePort(t *testing.T) {
	tests := []struct {
		options  *ServiceNodePortOpts
		expected *v1.Service
		expectErr   bool
	}{
		{
			options: &ServiceNodePortOpts{
				ServiceCommon{
	            Name:        "node-port-ok",
				TCP:         []string{"456", "321:908"},
				ClusterIP:   "",
				Type: v1.ServiceTypeNodePort,
				},
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "node-port-ok",
					Labels: map[string]string{"app": "node-port-ok"},
				},
				Spec: v1.ServiceSpec{Type: "NodePort",
					Ports: []v1.ServicePort{{Name: "456", Protocol: "TCP", Port: 456, TargetPort: intstr.IntOrString{Type: 0, IntVal: 456, StrVal: ""}, NodePort: 0},
						{Name: "321-908", Protocol: "TCP", Port: 321, TargetPort: intstr.IntOrString{Type: 0, IntVal: 908, StrVal: ""}, NodePort: 0}},
					Selector:  map[string]string{"app": "node-port-ok"},
					ClusterIP: "", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		service, err := tc.options.createService()
		if err != nil {
			t.Errorf("unexpected error:\n%#v\n", err)
			return
		}
		if tc.expectErr && err == nil {
		   continue
	    }
		if !apiequality.Semantic.DeepEqual(service, tc.expected) {
			t.Errorf("expected:\n%#v\ngot:\n%#v", tc.expected, service)
		}
	}
}

func TestCreateServiceLoadBalancer(t *testing.T) {
	tests := []struct {
		options  *ServiceLoadBalancerOpts
		expected *v1.Service
		expectErr   bool
	}{
		{
			options: &ServiceLoadBalancerOpts{
	            ServiceCommon{
	            Name:        "loadbalancer-ok",
				TCP:         []string{"456:9898"},
				ClusterIP:   "",
				Type: v1.ServiceTypeLoadBalancer,
	            },
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "loadbalancer-ok",
					Labels: map[string]string{"app": "loadbalancer-ok"},
				},
				Spec: v1.ServiceSpec{Type: "LoadBalancer",
					Ports:     []v1.ServicePort{{Name: "456-9898", Protocol: "TCP", Port: 456, TargetPort: intstr.IntOrString{Type: 0, IntVal: 9898, StrVal: ""}, NodePort: 0}},
					Selector:  map[string]string{"app": "loadbalancer-ok"},
					ClusterIP: "", ExternalIPs: []string(nil), LoadBalancerIP: ""},
			},
			expectErr: false,
		},	
	}

	for _, tc := range tests {
		service, err := tc.options.createService()
		if err != nil {
			t.Errorf("unexpected error:\n%#v\n", err)
			return
		}
		if tc.expectErr && err == nil {
		   continue
	    }
		if !apiequality.Semantic.DeepEqual(service, tc.expected) {
			t.Errorf("expected:\n%#v\ngot:\n%#v", tc.expected, service)
		}
	}
}

func TestCreateServiceExternalName(t *testing.T) {
	tests := []struct {
		options  *ServiceExternalNameOpts
		expected *v1.Service
		expectErr   bool
	}{
		{
			options: &ServiceExternalNameOpts{
	            ServiceCommon{
				Name:         "externalname-ok",
				Type:         v1.ServiceTypeExternalName,
				TCP:          []string{"123", "234:1234"},
				ClusterIP:    "",
				ExternalName: "test",
	            },
			},
			expected: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "externalname-ok",
					Labels: map[string]string{"app": "externalname-ok"},
				},
				//TODO: check why - not colon
				Spec: v1.ServiceSpec{Type: "ExternalName",
					Ports: []v1.ServicePort{{Name: "123", Protocol: "TCP", Port: 123, TargetPort: intstr.IntOrString{Type: 0, IntVal: 123, StrVal: ""}, NodePort: 0},
						{Name: "234-1234", Protocol: "TCP", Port: 234, TargetPort: intstr.IntOrString{Type: 0, IntVal: 1234, StrVal: ""}, NodePort: 0}},
					Selector:  map[string]string{"app": "externalname-ok"},
					ClusterIP: "", ExternalIPs: []string(nil), LoadBalancerIP: "", ExternalName: "test"},
			},
			expectErr: false,
		},	
	}

	for _, tc := range tests {
		service, err := tc.options.createService()
		if err != nil {
			t.Errorf("unexpected error:\n%#v\n", err)
			return
		}
		if tc.expectErr && err == nil {
		   continue
	    }
		if !apiequality.Semantic.DeepEqual(service, tc.expected) {
			t.Errorf("expected:\n%#v\ngot:\n%#v", tc.expected, service)
		}
	}
}

func TestParsePorts(t *testing.T) {
	tests := []struct {
		portString       string
		expectPort       int32
		expectTargetPort intstr.IntOrString
		expectErr        string
	}{
		{
			portString:       "3232",
			expectPort:       3232,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 3232},
		},
		{
			portString:       "1:65535",
			expectPort:       1,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 65535},
		},
		{
			portString:       "-5:1234",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 0},
			expectErr:        "parsing \"-5\": invalid syntax",
		},
		{
			portString:       "0:1234",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 1234},
		},
		{
			portString:       "5:65536",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 0},
			expectErr:        "must be between 1 and 65535, inclusive",
		},
		{
			portString:       "test-5:443",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 0},
			expectErr:        "invalid syntax",
		},
		{
			portString:       "5:test-443",
			expectPort:       5,
			expectTargetPort: intstr.IntOrString{Type: intstr.String, StrVal: "test-443"},
		},
		{
			portString:       "5:test*443",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 0},
			expectErr:        "must contain only alpha-numeric characters (a-z, 0-9), and hyphens (-)",
		},
		{
			portString:       "5:",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 0},
			expectErr:        "must contain at least one letter or number (a-z, 0-9)",
		},
		{
			portString:       "5:test--443",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 0},
			expectErr:        "must not contain consecutive hyphens",
		},
		{
			portString:       "5:test443-",
			expectPort:       0,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 0},
			expectErr:        "must not begin or end with a hyphen",
		},
		{
			portString:       "3232:1234:4567",
			expectPort:       3232,
			expectTargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 1234},
		},
	}

	for _, test := range tests {
		t.Run(test.portString, func(t *testing.T) {
			port, targetPort, err := parsePorts(test.portString)
			if len(test.expectErr) != 0 {
				if !strings.Contains(err.Error(), test.expectErr) {
					t.Errorf("parse ports string: %s. Expected err: %s, Got err: %v.", test.portString, test.expectErr, err)
				}
			}
			if !reflect.DeepEqual(targetPort, test.expectTargetPort) || port != test.expectPort {
				t.Errorf("parse ports string: %s. Expected port:%d, targetPort:%v, Got port:%d, targetPort:%v.", test.portString, test.expectPort, test.expectTargetPort, port, targetPort)
			}
		})
	}
}

func TestValidateService(t *testing.T) {
	tests := []struct {
		name      string
		s         ServiceCommon
		expectErr string
	}{
		{
			name: "validate-ok",
			s: ServiceCommon{
				Name:      "validate-ok",
				Type:      v1.ServiceTypeClusterIP,
				TCP:       []string{"123", "234:1234"},
				ClusterIP: "",
			},
		},
		{
			name: "Name-none",
			s: ServiceCommon{
				Type:      v1.ServiceTypeClusterIP,
				TCP:       []string{"123", "234:1234"},
				ClusterIP: "",
			},
			expectErr: "name must be specified",
		},
		{
			name: "Type-none",
			s: ServiceCommon{
				Name:      "validate-ok",
				TCP:       []string{"123", "234:1234"},
				ClusterIP: "",
			},
			expectErr: "type must be specified",
		},
		{
			name: "invalid-ClusterIPNone",
			s: ServiceCommon{
				Name:      "validate-ok",
				Type:      v1.ServiceTypeNodePort,
				TCP:       []string{"123", "234:1234"},
				ClusterIP: v1.ClusterIPNone,
			},
			expectErr: "ClusterIP=None can only be used with ClusterIP service type",
		},
		{
			name: "TCP-none",
			s: ServiceCommon{
				Name:      "validate-ok",
				Type:      v1.ServiceTypeClusterIP,
				ClusterIP: "",
			},
			expectErr: "at least one tcp port specifier must be provided",
		},
		{
			name: "invalid-ExternalName",
			s: ServiceCommon{
				Name:         "validate-ok",
				Type:         v1.ServiceTypeExternalName,
				TCP:          []string{"123", "234:1234"},
				ClusterIP:    "",
				ExternalName: "@oi:test",
			},
			expectErr: "invalid service external name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.s.Validate()
			if err != nil {
				if !strings.Contains(err.Error(), test.expectErr) {
					t.Errorf("validate:%s Expected err: %s, Got err: %v", test.name, test.expectErr, err)
				}
			}
			if err == nil && len(test.expectErr) != 0 {
				t.Errorf("validate:%s Expected success, Got err: %v", test.name, err)
			}
		})
	}
}


