class Macpine < Formula
  desc "Lightweight Alpine Virtual Machines on MacOS"
  homepage ""
  url "https://github.com/beringresearch/macpine/archive/refs/tags/v.01.tar.gz"
  sha256 "44454f832d28e91f4bc88bd55d5277ddfe046c54e0d2e68231689a960b7efe8e"
  license "Apache-2.0"

  depends_on "go" => :build
  depends_on "qemu"

  def install
    system "make", "all"

    bin.install Dir["_output/bin/*"]
    share.install Dir["_output/share/*"]
  end

  test do
    #assert_match "NAME STATUS SSH PORTS ARCH PID", shell_output("#{bin}/alpine list")
  end
end
