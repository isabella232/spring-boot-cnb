/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/v2/layers"
	"github.com/cloudfoundry/libcfbuildpack/v2/test"
	"github.com/cloudfoundry/spring-boot-cnb/cli"
	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCommand(t *testing.T) {
	spec.Run(t, "Spring Boot CLI Command", func(t *testing.T, when spec.G, it spec.S) {

		g := gomega.NewWithT(t)

		var f *test.BuildFactory

		it.Before(func() {
			f = test.NewBuildFactory(t)
		})

		when("NewCommand", func() {

			it("returns false when no groovy files", func() {
				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeFalse())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("returns true when jvm-application and groovy files", func() {
				test.CopyDirectory(t, filepath.Join("testdata", "valid_app"), f.Build.Application.Root)

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("ignores .groovy directories", func() {
				test.TouchFile(t, f.Build.Application.Root, "test.groovy", "test")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeFalse())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("rejects non-POGO, non-config files", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test.groovy"), "x")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeFalse())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("ignores logback files", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "ch", "qos", "logback", "test.groovy"), "class X {")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeFalse())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("detects POGO files", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test.groovy"), "class X {")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("detects config files", func() {
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "test.groovy"), "beans {")

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

			it("detects invalid .groovy files", func() {
				test.CopyFile(t, filepath.Join("testdata", "valid_app", "invalid.groovy"), filepath.Join(f.Build.Application.Root, "test.groovy"))

				_, ok, err := cli.NewCommand(f.Build)
				g.Expect(ok).To(gomega.BeTrue())
				g.Expect(err).NotTo(gomega.HaveOccurred())
			})

		})

		it("contributes command", func() {
			test.CopyDirectory(t, filepath.Join("testdata", "valid_app"), f.Build.Application.Root)

			c, ok, err := cli.NewCommand(f.Build)
			g.Expect(ok).To(gomega.BeTrue())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			g.Expect(c.Contribute()).To(gomega.Succeed())

			layer := f.Build.Layers.Layer("command")
			g.Expect(layer).To(test.HaveLayerMetadata(false, false, true))
			g.Expect(layer).To(test.HaveAppendLaunchEnvironment("GROOVY_FILES", strings.Join([]string{
				"",
				filepath.Join(f.Build.Application.Root, "directory", "pogo_4.groovy"),
				filepath.Join(f.Build.Application.Root, "invalid.groovy"),
				filepath.Join(f.Build.Application.Root, "pogo_1.groovy"),
				filepath.Join(f.Build.Application.Root, "pogo_2.groovy"),
				filepath.Join(f.Build.Application.Root, "pogo_3.groovy"),
			}, " ")))

			command := "spring run -cp $CLASSPATH $GROOVY_FILES"
			g.Expect(f.Build.Layers).To(test.HaveApplicationMetadata(layers.Metadata{
				Processes: []layers.Process{
					{Type: "spring-boot-cli", Command: command},
					{Type: "task", Command: command},
					{Type: "web", Command: command},
				},
			}))
		})
	}, spec.Report(report.Terminal{}))
}
