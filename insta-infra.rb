class InstaInfra < Formula
  desc "Spin up any service straight away on your local laptop"
  homepage "https://github.com/data-catering/insta-infra"
  url "https://github.com/data-catering/insta-infra/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256_AFTER_FIRST_RELEASE"
  license "MIT"

  depends_on "docker" => :recommended
  depends_on "docker-compose" => :recommended

  def install
    bin.install "run.sh" => "insta"
    prefix.install "docker-compose.yaml"
    prefix.install "docker-compose-persist.yaml"
    prefix.install "data"
    prefix.install "README.md"
  end

  def caveats
    <<~EOS
      To use insta-infra, run:
        insta <service-name>
      
      For help, run:
        insta help
      
      Persisted data will be stored in:
        #{prefix}/data/<service>/persist
    EOS
  end

  test do
    system "#{bin}/insta", "-l"
  end
end 