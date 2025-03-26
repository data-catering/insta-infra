%global __go %{__gobuild}

Name:           insta
Version:        0.1.0
Release:        1%{?dist}
Summary:        A simple, fast CLI tool for spinning up data infrastructure services
License:        MIT
URL:            https://github.com/data-catering/insta-infra
Source0:        insta-%{version}.tar.gz
BuildRequires:  golang >= 1.20
Requires:       docker >= 20.10 || podman >= 3.0

%description
Insta is a command-line tool that makes it easy to run data infrastructure
services using Docker or Podman. It provides pre-configured services
with optional data persistence and simple connection management.

%prep
%autosetup

%build
%gobuild -o %{name} ./cmd/insta

%install
mkdir -p %{buildroot}%{_bindir}
install -m 755 %{name} %{buildroot}%{_bindir}/

mkdir -p %{buildroot}%{_mandir}/man1
install -m 644 docs/man/%{name}.1 %{buildroot}%{_mandir}/man1/

%files
%{_bindir}/%{name}
%{_mandir}/man1/%{name}.1*

%changelog
* Mon Mar 25 2024 Peter Flook <peter.flook@data.catering> - 0.1.0-1
- Initial release
- Support for Docker and Podman
- Pre-configured data infrastructure services
- Optional data persistence
- Simple service connection management 