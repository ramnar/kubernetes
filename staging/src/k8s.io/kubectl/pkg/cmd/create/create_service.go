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
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/generate"
	generateversioned "k8s.io/kubectl/pkg/generate/versioned"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// NewCmdCreateService is a macro command to create a new service
func NewCmdCreateService(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc"},
		Short:   i18n.T("Create a service using specified subcommand."),
		Long:    i18n.T("Create a service using specified subcommand."),
		Run:     cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}
	cmd.AddCommand(NewCmdCreateServiceClusterIP(f, ioStreams))
	cmd.AddCommand(NewCmdCreateServiceNodePort(f, ioStreams))
	cmd.AddCommand(NewCmdCreateServiceLoadBalancer(f, ioStreams))
	cmd.AddCommand(NewCmdCreateServiceExternalName(f, ioStreams))

	return cmd
}

var (
	serviceClusterIPLong = templates.LongDesc(i18n.T(`
    Create a ClusterIP service with the specified name.`))

	serviceClusterIPExample = templates.Examples(i18n.T(`
    # Create a new ClusterIP service named my-cs
    kubectl create service clusterip my-cs --tcp=5678:8080

    # Create a new ClusterIP service named my-cs (in headless mode)
    kubectl create service clusterip my-cs --clusterip="None"`))
)

func addPortFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("tcp", []string{}, "Port pairs can be specified as '<port>:<targetPort>'.")
}

// ServiceClusterIPOpts holds the options for 'create service clusterip' sub command
type ServiceClusterIPOpts struct {
	// PrintFlags holds options necessary for obtaining a printer
	PrintFlags *genericclioptions.PrintFlags
	PrintObj   func(obj runtime.Object) error
	// Name of resource being created
	Name             string
	DefaultName      string
	Selector         string
	// Port will be used if a user specifies --port OR the exposed object
	// has one port
	Port             string
	// Ports will be used if a user doesn't specify --port AND the
	// exposed object has multiple ports
	Ports            string
	Labels           string
	ExternalIP       string
	LoadBalancerIP   string
	Type             string
	Protocol         string
	// protocols will be used to keep port-protocol mapping derived from
	// exposed object
	Protocols        string
	// alias of TargetPort
	ContainerPort    string
	TargetPort       string
	PortName         string
	SessionAffinity  string
	ClusterIP        string
	
	DryRunStrategy   cmdutil.DryRunStrategy
	DryRunVerifier   *resourcecli.DryRunVerifier
	CreateAnnotation bool
	FieldManager     string

	Namespace        string
	EnforceNamespace bool

	Mapper meta.RESTMapper
	Client *coreclient.CoreV1Client

	genericclioptions.IOStreams
}

// NewServiceClusterIPOpts creates a new *ServiceClusterIPOpts with sane defaults
func NewServiceClusterIPOpts(ioStreams genericclioptions.IOStreams) *ServiceClusterIPOpts {
	return &ServiceClusterIPOpts{
		PrintFlags: genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		IOStreams:  ioStreams,
	}
}

// NewCmdCreateServiceClusterIP is a command to create a ClusterIP service
func NewCmdCreateServiceClusterIP(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := NewServiceClusterIPOpts(ioStreams)

	cmd := &cobra.Command{
		Use:                   "clusterip NAME [--tcp=<port>:<targetPort>] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create a ClusterIP service."),
		Long:                  serviceClusterIPLong,
		Example:               serviceClusterIPExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

    o.PrintFlags.AddFlags(cmd)
    
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)
	addPortFlags(cmd)
	cmd.Flags().String("clusterip", "", i18n.T("Assign your own ClusterIP or set to 'None' for a 'headless' service (no loadbalancing)."))
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-create")
	return cmd
}

func errUnsupportedGenerator(cmd *cobra.Command, generatorName string) error {
	return cmdutil.UsageErrorf(cmd, "Generator %s not supported. ", generatorName)
}

// Complete completes all the required options
func (o *ServiceClusterIPOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}


	restConfig, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	o.Client, err = coreclient.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	o.CreateAnnotation = cmdutil.GetFlagBool(cmd, cmdutil.ApplyAnnotationsFlag)

	o.DryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	if err != nil {
		return err
	}
	dynamicClient, err := f.DynamicClient()
	if err != nil {
		return err
	}
	discoveryClient, err := f.ToDiscoveryClient()
	if err != nil {
		return err
	}
	o.DryRunVerifier = resourcecli.NewDryRunVerifier(dynamicClient, discoveryClient)

	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.DryRunStrategy)

	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	o.PrintObj = func(obj runtime.Object) error {
		return printer.PrintObj(obj, o.Out)
	}

	return nil
}

// Validate checks to the ServiceClusterIPOpts to see if there is sufficient information run the command.
func (o *ServiceClusterIPOpts) Validate() error {
	if len(o.Name) == 0 {
		return fmt.Errorf("name must be specified")
	}
	
	if len(o.Selector) == 0 {
		return fmt.Errorf("'selector' is a required parameter")
	}
	return nil
}

// Run calls the CreateSubcommandOptions.Run in ServiceClusterIPOpts instance
func (o *ServiceClusterIPOpts) Run() error {
	service, err := o.createService()
	if err != nil {
		return err
	}

	if err := util.CreateOrUpdateAnnotation(o.CreateAnnotation, service, scheme.DefaultJSONEncoder()); err != nil {
		return err
	}

	if o.DryRunStrategy != cmdutil.DryRunClient {
		createOptions := metav1.CreateOptions{}
		if o.FieldManager != "" {
			createOptions.FieldManager = o.FieldManager
		}
		if o.DryRunStrategy == cmdutil.DryRunServer {
			if err := o.DryRunVerifier.HasSupport(service.GroupVersionKind()); err != nil {
				return err
			}
			createOptions.DryRun = []string{metav1.DryRunAll}
		}
		service, err = o.Client.Service(o.Namespace).Create(context.TODO(), service, createOptions)
		if err != nil {
			return fmt.Errorf("failed to create service: %v", err)
		}
	}
	return o.PrintObj(service)
}

func (o *ServiceClusterIPOpts) createService() (*corev1.Service, error) {
	namespace := ""
	if o.EnforceNamespace {
		namespace = o.Namespace
	}
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   o.name,
			Namespace: namespace,
			Labels: o.labels,
		},
		Spec: &corev1.ServiceSpec{
			Selector: o.selector,
			Ports:    o.ports,
		},
	}

	targetPortString := o.TargetPort
	if len(targetPortString) == 0 {
		targetPortString = o.ContainerPort
	}
	if len(targetPortString) > 0 {
		var targetPort intstr.IntOrString
		if portNum, err := strconv.Atoi(targetPortString); err != nil {
			targetPort = intstr.FromString(targetPortString)
		} else {
			targetPort = intstr.FromInt(portNum)
		}
		// Use the same target-port for every port
		for i := range service.Spec.Ports {
			service.Spec.Ports[i].TargetPort = targetPort
		}
	} else {
		// If --target-port or --container-port haven't been specified, this
		// should be the same as Port
		for i := range service.Spec.Ports {
			port := service.Spec.Ports[i].Port
			service.Spec.Ports[i].TargetPort = intstr.FromInt(int(port))
		}
	}
	if len(o.ExternalIP) > 0 {
		service.Spec.ExternalIPs = []string{o.ExternalIP}
	}
	if len(o.Type) != 0 {
		service.Spec.Type = &corev1.ServiceType(o.Type)
	}
	if service.Spec.Type == &corev1.ServiceTypeLoadBalancer {
		service.Spec.LoadBalancerIP = o.LoadBalancerIP
	}
	if len(o.SessionAffinity) != 0 {
		switch &corev1.ServiceAffinity(o.SessionAffinity) {
		case &corev1.ServiceAffinityNone:
			service.Spec.SessionAffinity = &corev1.ServiceAffinityNone
		case &corev1.ServiceAffinityClientIP:
			service.Spec.SessionAffinity = &corev1.ServiceAffinityClientIP
		default:
			return nil, fmt.Errorf("unknown session affinity: %s", o.SessionAffinity)
		}
	}
	if len(o.ClusterIP) != 0 {
		if o.ClusterIP == "None" {
			service.Spec.ClusterIP = &corev1.ClusterIPNone
		} else {
			service.Spec.ClusterIP = o.ClusterIP
		}
	}

	return service, nil
}

var (
	serviceNodePortLong = templates.LongDesc(i18n.T(`
    Create a NodePort service with the specified name.`))

	serviceNodePortExample = templates.Examples(i18n.T(`
    # Create a new NodePort service named my-ns
    kubectl create service nodeport my-ns --tcp=5678:8080`))
)

// ServiceNodePortOpts holds the options for 'create service nodeport' sub command
type ServiceNodePortOpts struct {
	CreateSubcommandOptions *CreateSubcommandOptions
}

// NewCmdCreateServiceNodePort is a macro command for creating a NodePort service
func NewCmdCreateServiceNodePort(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	options := &ServiceNodePortOpts{
		CreateSubcommandOptions: NewCreateSubcommandOptions(ioStreams),
	}

	cmd := &cobra.Command{
		Use:                   "nodeport NAME [--tcp=port:targetPort] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create a NodePort service."),
		Long:                  serviceNodePortLong,
		Example:               serviceNodePortExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(f, cmd, args))
			cmdutil.CheckErr(options.Run())
		},
	}

	options.CreateSubcommandOptions.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, generateversioned.ServiceNodePortGeneratorV1Name)
	cmd.Flags().Int("node-port", 0, "Port used to expose the service on each node in a cluster.")
	cmdutil.AddFieldManagerFlagVar(cmd, &options.CreateSubcommandOptions.FieldManager, "kubectl-create")
	addPortFlags(cmd)
	return cmd
}

// Complete completes all the required options
func (o *ServiceNodePortOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	var generator generate.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case generateversioned.ServiceNodePortGeneratorV1Name:
		generator = &generateversioned.ServiceCommonGeneratorV1{
			Name:      name,
			TCP:       cmdutil.GetFlagStringSlice(cmd, "tcp"),
			Type:      v1.ServiceTypeNodePort,
			ClusterIP: "",
			NodePort:  cmdutil.GetFlagInt(cmd, "node-port"),
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}

	return o.CreateSubcommandOptions.Complete(f, cmd, args, generator)
}

// Run calls the CreateSubcommandOptions.Run in ServiceNodePortOpts instance
func (o *ServiceNodePortOpts) Run() error {
	return o.CreateSubcommandOptions.Run()
}

var (
	serviceLoadBalancerLong = templates.LongDesc(i18n.T(`
    Create a LoadBalancer service with the specified name.`))

	serviceLoadBalancerExample = templates.Examples(i18n.T(`
    # Create a new LoadBalancer service named my-lbs
    kubectl create service loadbalancer my-lbs --tcp=5678:8080`))
)

// ServiceLoadBalancerOpts holds the options for 'create service loadbalancer' sub command
type ServiceLoadBalancerOpts struct {
	CreateSubcommandOptions *CreateSubcommandOptions
}

// NewCmdCreateServiceLoadBalancer is a macro command for creating a LoadBalancer service
func NewCmdCreateServiceLoadBalancer(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	options := &ServiceLoadBalancerOpts{
		CreateSubcommandOptions: NewCreateSubcommandOptions(ioStreams),
	}

	cmd := &cobra.Command{
		Use:                   "loadbalancer NAME [--tcp=port:targetPort] [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create a LoadBalancer service."),
		Long:                  serviceLoadBalancerLong,
		Example:               serviceLoadBalancerExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(f, cmd, args))
			cmdutil.CheckErr(options.Run())
		},
	}

	options.CreateSubcommandOptions.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, generateversioned.ServiceLoadBalancerGeneratorV1Name)
	cmdutil.AddFieldManagerFlagVar(cmd, &options.CreateSubcommandOptions.FieldManager, "kubectl-create")
	addPortFlags(cmd)
	return cmd
}

// Complete completes all the required options
func (o *ServiceLoadBalancerOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	var generator generate.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case generateversioned.ServiceLoadBalancerGeneratorV1Name:
		generator = &generateversioned.ServiceCommonGeneratorV1{
			Name:      name,
			TCP:       cmdutil.GetFlagStringSlice(cmd, "tcp"),
			Type:      v1.ServiceTypeLoadBalancer,
			ClusterIP: "",
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}

	return o.CreateSubcommandOptions.Complete(f, cmd, args, generator)
}

// Run calls the CreateSubcommandOptions.Run in ServiceLoadBalancerOpts instance
func (o *ServiceLoadBalancerOpts) Run() error {
	return o.CreateSubcommandOptions.Run()
}

var (
	serviceExternalNameLong = templates.LongDesc(i18n.T(`
	Create an ExternalName service with the specified name.

	ExternalName service references to an external DNS address instead of
	only pods, which will allow application authors to reference services
	that exist off platform, on other clusters, or locally.`))

	serviceExternalNameExample = templates.Examples(i18n.T(`
	# Create a new ExternalName service named my-ns
	kubectl create service externalname my-ns --external-name bar.com`))
)

// ServiceExternalNameOpts holds the options for 'create service externalname' sub command
type ServiceExternalNameOpts struct {
	CreateSubcommandOptions *CreateSubcommandOptions
}

// NewCmdCreateServiceExternalName is a macro command for creating an ExternalName service
func NewCmdCreateServiceExternalName(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	options := &ServiceExternalNameOpts{
		CreateSubcommandOptions: NewCreateSubcommandOptions(ioStreams),
	}

	cmd := &cobra.Command{
		Use:                   "externalname NAME --external-name external.name [--dry-run=server|client|none]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Create an ExternalName service."),
		Long:                  serviceExternalNameLong,
		Example:               serviceExternalNameExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(f, cmd, args))
			cmdutil.CheckErr(options.Run())
		},
	}

	options.CreateSubcommandOptions.PrintFlags.AddFlags(cmd)

	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, generateversioned.ServiceExternalNameGeneratorV1Name)
	addPortFlags(cmd)
	cmd.Flags().String("external-name", "", i18n.T("External name of service"))
	cmd.MarkFlagRequired("external-name")
	cmdutil.AddFieldManagerFlagVar(cmd, &options.CreateSubcommandOptions.FieldManager, "kubectl-create")
	return cmd
}

// Complete completes all the required options
func (o *ServiceExternalNameOpts) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}

	var generator generate.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case generateversioned.ServiceExternalNameGeneratorV1Name:
		generator = &generateversioned.ServiceCommonGeneratorV1{
			Name:         name,
			Type:         v1.ServiceTypeExternalName,
			ExternalName: cmdutil.GetFlagString(cmd, "external-name"),
			ClusterIP:    "",
		}
	default:
		return errUnsupportedGenerator(cmd, generatorName)
	}

	return o.CreateSubcommandOptions.Complete(f, cmd, args, generator)
}

// Run calls the CreateSubcommandOptions.Run in ServiceExternalNameOpts instance
func (o *ServiceExternalNameOpts) Run() error {
	return o.CreateSubcommandOptions.Run()
}
