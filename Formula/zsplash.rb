class Zsplash < Formula
  desc "Lightweight CLI utility for colorful, expressive Zsh & terminal text (includes zsp alias)"
  homepage "https://github.com/acorrow/zsplash"
  url "https://github.com/acorrow/zsplash/archive/refs/tags/v0.3.0.tar.gz"
  sha256 "fa01a91ca351ac56fc1152de7b5074375860415408f61ced5735e10064f34b71"
  license "MIT"
  head "https://github.com/acorrow/zsplash.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}")
    bin.install_symlink "zsplash" => "zsp"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/zsplash --version")
    assert_match version.to_s, shell_output("#{bin}/zsp --version")
  end
end