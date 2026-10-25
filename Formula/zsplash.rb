class Zsplash < Formula
  desc "Lightweight CLI utility for colorful, expressive Zsh & terminal text (includes zsp alias)"
  homepage "https://github.com/acorrow/zsplash"
  url "https://github.com/acorrow/zsplash/archive/refs/tags/v0.3.1.tar.gz"
  sha256 "ca97fc640c3c2e9517c69e87d6ebe4ae8a0ffd41b3f712657c1dde7d0104f944"
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