// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cli implements the commandline interface for Pixiecore.
package cli // import "go.universe.tf/netboot/pixiecore/cli"

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.universe.tf/netboot/pixiecore"
)

// Ipxe is the set of ipxe binaries for supported firmwares.
//
// Can be set externally before calling CLI(), and set/extended by
// commandline processing in CLI().
var Ipxe = map[pixiecore.Firmware][]byte{}

// CLI runs the Pixiecore commandline.
//
// This function always exits back to the OS when finished.
func CLI() {
	if v1compatCLI() {
		return
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	os.Exit(0)
}

// This represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pixiecore",
	Short: "All-in-one network booting",
	Long:  `Pixiecore is a tool to make network booting easy.`,
}

func initConfig() {
	viper.SetEnvPrefix("pixiecore")
	viper.AutomaticEnv() // read in environment variables that match
}

func fatalf(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}

func todo(msg string, args ...interface{}) {
	fatalf("TODO: "+msg, args...)
}

func serverConfigFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("debug", "d", false, "Log more things that aren't directly related to booting a recognized client")
	cmd.Flags().BoolP("log-timestamps", "t", false, "Add a timestamp to each log line")
	cmd.Flags().IPP("listen-addr", "l", net.IPv4zero, "IPv4 address to listen on")
	cmd.Flags().IntP("port", "p", 80, "Port to listen on for HTTP")
	cmd.Flags().String("ipxe-bios", "", "path to an iPXE binary for BIOS/UNDI")
	cmd.Flags().String("ipxe-efi32", "", "path to an iPXE binary for 32-bit UEFI")
	cmd.Flags().String("ipxe-efi64", "", "path to an iPXE binary for 64-bit UEFI")
}

func mustFile(path string) []byte {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		fatalf("couldn't read file %q: %s", path, err)
	}

	return bs
}

func serverFromFlags(cmd *cobra.Command) *pixiecore.Server {
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	timestamps, err := cmd.Flags().GetBool("log-timestamps")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	addr, err := cmd.Flags().GetIP("listen-addr")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	httpPort, err := cmd.Flags().GetInt("port")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	ipxeBios, err := cmd.Flags().GetString("ipxe-bios")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	ipxeEFI32, err := cmd.Flags().GetString("ipxe-efi32")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}
	ipxeEFI64, err := cmd.Flags().GetString("ipxe-efi64")
	if err != nil {
		fatalf("Error reading flag: %s", err)
	}

	if addr != nil && addr.To4() == nil {
		fatalf("Listen address must be IPv4")
	}
	if httpPort <= 0 {
		fatalf("HTTP port must be >0")
	}

	ret := &pixiecore.Server{
		Ipxe:     map[pixiecore.Firmware][]byte{},
		Log:      logWithStdFmt,
		HTTPPort: httpPort,
	}
	for fwtype, bs := range Ipxe {
		ret.Ipxe[fwtype] = bs
	}
	if ipxeBios != "" {
		ret.Ipxe[pixiecore.FirmwareX86PC] = mustFile(ipxeBios)
	}
	if ipxeEFI32 != "" {
		ret.Ipxe[pixiecore.FirmwareEFI32] = mustFile(ipxeEFI32)
	}
	if ipxeEFI64 != "" {
		ret.Ipxe[pixiecore.FirmwareEFI64] = mustFile(ipxeEFI64)
	}

	if timestamps {
		ret.Log = logWithStdLog
	}
	if debug {
		ret.Debug = ret.Log
	}
	if addr != nil {
		ret.Address = addr.String()
	}

	return ret
}
