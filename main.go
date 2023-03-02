package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/pterm/pcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var template_path_cyclonedds string
var template_path_fastdds_simple string
var template_path_fastdds_ds_client string
var template_path_fastdds_ds_server string
var dds_discovery_server_port string

var version string = "unknown"

//go:embed dds-templates/cyclonedds-simple.xml
//go:embed dds-templates/fastdds-simple.xml
//go:embed dds-templates/fastdds-ds-server.xml
//go:embed dds-templates/fastdds-ds-client.xml
var f embed.FS
var husarnet_temp_dir string

func main_loop() {
	default_config_cyclonedds_simple, _ := f.ReadFile("dds-templates/cyclonedds-simple.xml")
	default_config_fastdds_simple, _ := f.ReadFile("dds-templates/fastdds-simple.xml")
	default_config_fastdds_ds_server, _ := f.ReadFile("dds-templates/fastdds-ds-server.xml")
	default_config_fastdds_ds_client, _ := f.ReadFile("dds-templates/fastdds-ds-client.xml")

	if HusarnetPresent() {
		var output_xml string
		var input_xml string
		var output_xml_path string

		myos := runtime.GOOS
		fmt.Printf("Host OS: %s\n", myos)
		switch myos {
		case "linux":
			husarnet_temp_dir = "/var/tmp/husarnet-dds"
		default:
			husarnet_temp_dir = os.TempDir() + "/husarnet-dds"
		}

		fmt.Println("Temporary directory:", husarnet_temp_dir)

		// Prepare a config for Discovery Server (server) in all cases
		input_xml = string(default_config_fastdds_ds_server)

		// check if non-default DDS config file exists
		if _, err := os.Stat(template_path_fastdds_ds_server); err == nil {
			input_xml_bytes, _ := ioutil.ReadFile(template_path_fastdds_ds_server)
			input_xml = string(input_xml_bytes)
		}

		output_xml = strings.Replace(input_xml, "$HOST_IPV6", GetOwnHusarnetIPv6(), -1)

		// Read the port for the Discovery Server "server" config
		dds_discovery_server_port, ok := os.LookupEnv("DISCOVERY_SERVER_PORT")
		if ok {
			fmt.Println("DISCOVERY_SERVER_PORT:", dds_discovery_server_port)
		} else {
			dds_discovery_server_port = "11811"
		}

		output_xml = strings.Replace(output_xml, "$DISCOVERY_SERVER_PORT", dds_discovery_server_port, -1)

		output_xml_path = husarnet_temp_dir + "/fastdds-ds-server.xml"

		// Create necessary directories for Discovery Server (server) xml config
		dir := filepath.Dir(output_xml_path)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Printf("Err: can not create \"%s\" path", dir)
			os.Exit(1)
		}

		ioutil.WriteFile(output_xml_path, []byte(output_xml), 0644)

		// Check the RMW_IMPLEMENTATION env
		rmw_implementation, ok := os.LookupEnv("RMW_IMPLEMENTATION")
		if ok {
			fmt.Println("RMW_IMPLEMENTATION:", rmw_implementation)
		} else {
			fmt.Println("RMW_IMPLEMENTATION is not set.")
			return
		}

		if rmw_implementation == "rmw_cyclonedds_cpp" {
			// default config
			input_xml = string(default_config_cyclonedds_simple)

			// check if non-default DDS config file exists
			if _, err := os.Stat(template_path_cyclonedds); err == nil {
				input_xml_bytes, _ := ioutil.ReadFile(template_path_cyclonedds)
				input_xml = string(input_xml_bytes)
			}

			// prepare XML files with Husarnet hosts
			output_xml = ParseCycloneDDSSimple(input_xml)

			// defaul output path
			output_xml_path = husarnet_temp_dir + "/husarnet-cyclonedds.xml"

			// check whether env to set non-default path is set
			cyclonedds_uri, ok := os.LookupEnv("CYCLONEDDS_URI")
			if ok {
				fmt.Println("CYCLONEDDS_URI:", cyclonedds_uri)
				if strings.Contains(cyclonedds_uri, "husarnet") {
					output_xml_path = strings.Split(cyclonedds_uri, "file://")[1]
				}
			}
		}

		if rmw_implementation == "rmw_fastrtps_cpp" {

			// Load the appriopriate XML default config
			ros_discovery_server, is_ds_client := os.LookupEnv("ROS_DISCOVERY_SERVER")
			if is_ds_client {
				fmt.Println("ROS_DISCOVERY_SERVER:", ros_discovery_server)
				input_xml = string(default_config_fastdds_ds_client)
			} else {
				input_xml = string(default_config_fastdds_simple)
			}

			// check if non-default DDS config file exists
			if _, err := os.Stat(template_path_fastdds_simple); err == nil {
				input_xml_bytes, _ := ioutil.ReadFile(template_path_fastdds_simple)
				input_xml = string(input_xml_bytes)
			}

			// prepare XML files with Husarnet hosts
			if is_ds_client {
				var ds_server_addr string
				var ds_server_port string

				// check whether IPv6 address is provided instead of hostname
				ipv6 := strings.Split(ros_discovery_server, ":")

				if ipv6[0] == "[fc94" {
					// IPv6 hostname is provided
					parts := strings.Split(ros_discovery_server, "]:")

					if len(parts) == 1 {
						fmt.Println("Error: Invalid string format")
						os.Exit(1)
					}

					ds_server_addr = strings.Trim(parts[0], "[")
					ds_server_port = parts[1]
					output_xml = strings.Replace(input_xml, "$DISCOVERY_SERVER_IPV6", ds_server_addr, 1)
				} else {
					// normal hostname is provided
					ds_server_addr = strings.Split(ros_discovery_server, ":")[0]
					ds_server_port = strings.Split(ros_discovery_server, ":")[1]
					output_xml = strings.Replace(input_xml, "$DISCOVERY_SERVER_IPV6", GetHostIPv6(ds_server_addr), 1)
				}

				output_xml = strings.Replace(output_xml, "$DISCOVERY_SERVER_PORT", ds_server_port, 1)
				output_xml = strings.Replace(output_xml, "$HOST_IPV6", GetOwnHusarnetIPv6(), -1)
			} else {
				output_xml = ParseFastDDSSimple(input_xml)
			}

			// defaul output path
			output_xml_path = husarnet_temp_dir + "/husarnet-fastdds.xml"

			// check whether env to set non-default path is set
			fastrtps_default_profiles_file, ok := os.LookupEnv("FASTRTPS_DEFAULT_PROFILES_FILE")
			if ok {
				fmt.Println("FASTRTPS_DEFAULT_PROFILES_FILE:", fastrtps_default_profiles_file)
				if strings.Contains(fastrtps_default_profiles_file, "husarnet") {
					output_xml_path = fastrtps_default_profiles_file
				}
			}

		}

		// Create necessary directories
		dir = filepath.Dir(output_xml_path)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Printf("Err: can not create \"%s\" path", dir)
			os.Exit(1)
		}

		ioutil.WriteFile(output_xml_path, []byte(output_xml), 0644)
		fmt.Printf("DDS config saved here: \"%s\"", output_xml_path)
	} else {
		fmt.Println("can't reach Husarnet client API")
		os.Exit(1)
	}
}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start the service logic in a goroutine
	go p.run()
	return nil
}

func (p *program) run() {
	// Your service logic here
	fmt.Println("Service loop")
	for {
		fmt.Println("===============================")
		main_loop()
		time.Sleep(5 * time.Second)
	}
}

func (p *program) Stop(s service.Service) error {
	// Your service logic here
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "husarnet-dds",
	Short: "Create DDS config for Husarnet automatically",
	Long:  `Create DDS config for Husarnet automatically`,
	Example: `
husarnet-dds singleshot
husarnet-dds install $USER
husarnet-dds start
husarnet-dds stop
husarnet-dds uninstall
	`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	// Fetch user interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pterm.Warning.Println("user interrupt")
		pcli.CheckForUpdates()
		os.Exit(0)
	}()

	// Execute cobra
	if err := rootCmd.Execute(); err != nil {
		pcli.CheckForUpdates()
		os.Exit(1)
	}

	pcli.CheckForUpdates()
}

func main() {

	userName := "root"

	prg := &program{}

	svcConfig := &service.Config{
		Name:        "husarnet-dds",
		DisplayName: "Husarnet DDS Configurator",
		Description: "Creating a DDS config for Husarion VPN",
		Arguments:   []string{"daemon"},
		UserName:    userName,
	}
	s, err := service.New(prg, svcConfig)

	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	var envs []string

	// Define your CLI commands and flags here
	installCommand := &cobra.Command{
		Use:   "install",
		Short: "Install the service",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				userName = args[0]
			}

			envvars := make(map[string]string)

			for i, env := range envs {
				fmt.Printf("env[%d]: %s\n", i, env)
				key := strings.Split(env, "=")[0]
				value := strings.Split(env, "=")[1]

				envvars[key] = value
			}

			options := service.KeyValue{}

			if runtime.GOOS == "windows" {
				userName = "LocalSystem"
			}

			if runtime.GOOS == "darwin" {
				options["LogDirectory"] = "/tmp/"
			}

			fmt.Println("username:", userName)

			svcConfig = &service.Config{
				Name:        "husarnet-dds",
				DisplayName: "Husarnet DDS Configurator",
				Description: "Creating a DDS config for Husarion VPN",
				Arguments:   []string{"daemon"},
				UserName:    userName,
				EnvVars:     envvars,
				Option:      options,
			}
			s, err := service.New(prg, svcConfig)

			if err != nil {
				log.Fatalf("Failed to create service: %v", err)
			}

			err = s.Install()
			if err != nil {
				log.Fatalf("Failed to install service: %v", err)
			}
			fmt.Println("Service installed.")
		},
	}

	installCommand.Flags().StringArrayVarP(&envs, "env", "e",
		[]string{
			"RMW_IMPLEMENTATION=rmw_fastrtps_cpp",
			"FASTRTPS_DEFAULT_PROFILES_FILE=" + husarnet_temp_dir + "/fastdds.xml"},
		"environment variables for the service")

	uninstallCommand := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the service",
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Uninstall()
			if err != nil {
				log.Fatalf("Failed to uninstall service: %v", err)
			}
			fmt.Println("Service removed.")
		},
	}

	startCommand := &cobra.Command{
		Use:   "start",
		Short: "Start the service",
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Start()
			if err != nil {
				log.Fatalf("Failed to start service: %v", err)
			}
			fmt.Println("Service started.")
		},
	}

	stopCommand := &cobra.Command{
		Use:   "stop",
		Short: "Stop the service",
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Stop()
			if err != nil {
				log.Fatalf("Failed to stop service: %v", err)
			}
			fmt.Println("Service stopped.")
		},
	}

	daemonCommand := &cobra.Command{
		Use:   "daemon",
		Short: "Run the program in the inifinite loop",
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Run()
			if err != nil {
				log.Fatalf("Failed to run service: %v", err)
			}
		},
	}

	singleShotCommand := &cobra.Command{
		Use:   "singleshot",
		Short: "Run the program only once (not as a service)",
		Run: func(cmd *cobra.Command, args []string) {
			main_loop()
		},
	}

	rootCmd.AddCommand(installCommand)
	rootCmd.AddCommand(uninstallCommand)
	rootCmd.AddCommand(startCommand)
	rootCmd.AddCommand(stopCommand)
	rootCmd.AddCommand(daemonCommand)
	rootCmd.AddCommand(singleShotCommand)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	pcli.SetRepo("husarnet/husarnet-dds")
	pcli.DisableUpdateChecking = true
	pcli.SetRootCmd(rootCmd)
	pcli.Setup()

	Execute()
}
