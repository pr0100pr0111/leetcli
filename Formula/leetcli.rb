class Leetcli < Formula
  desc "Polished terminal dashboard for your LeetCode profile"
  homepage "https://github.com/yourname/leetcli"
  version "VERSION_PLACEHOLDER"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yourname/leetcli/releases/download/v#{version}/leetcli-darwin-arm64.tar.gz"
      sha256 "SHA256_PLACEHOLDER_ARM64"
    else
      url "https://github.com/yourname/leetcli/releases/download/v#{version}/leetcli-darwin-amd64.tar.gz"
      sha256 "SHA256_PLACEHOLDER_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yourname/leetcli/releases/download/v#{version}/leetcli-linux-arm64.tar.gz"
      sha256 "SHA256_PLACEHOLDER_LINUX_ARM64"
    else
      url "https://github.com/yourname/leetcli/releases/download/v#{version}/leetcli-linux-amd64.tar.gz"
      sha256 "SHA256_PLACEHOLDER_LINUX_AMD64"
    end
  end

  def install
    bin.install "leetcli"
  end

  test do
    assert_match "leetcli", shell_output("#{bin}/leetcli --help 2>&1", 1)
  end
end