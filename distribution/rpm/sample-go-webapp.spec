#
# spec file for package sample-go-webapp
#
# Copyright (c) 2025 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#


%define webdir %{_datadir}/%{name}/web
%if 0%{?suse_version} < 1600
%define apparmor_profilesdir %{_sysconfdir}/apparmor.d
%endif
Name:           sample-go-webapp
Version:        0
Release:        0
Summary:        Web application
License:        GPL-3.0-or-later
Group:          Productivity/Security
URL:            https://github.com/tacerus/sample-go-webapp
Source0:        %{name}-%{version}.tar.zst
Source1:        vendor.tar.gz
BuildRequires:  apparmor-rpm-macros
BuildRequires:  go >= 1.25
BuildRequires:  systemd-rpm-macros
BuildRequires:  zstd

%description
Dummy web application with OIDC login.

%prep
%autosetup -a1

%build
go build -buildmode=pie -mod=vendor ./cmd/webapp

%install
install -d \
	%{buildroot}%{_bindir} \
	%{buildroot}%{_sbindir} \
	%{buildroot}%{_sysconfdir} \
	%{buildroot}%{_unitdir} \
	%{buildroot}%{apparmor_profilesdir} \
	%{buildroot}%{webdir} \
%{nil}

install webapp %{buildroot}%{_bindir}/%{name}

sed -E 's?("AssetDir": ).*,$?\1"%{webdir}"?' config.json.example > %{buildroot}%{_sysconfdir}/%{name}.json

cp -r web/assets %{buildroot}%{webdir}

#install -m0644 distribution/apparmor/#{name}.apparmor %{buildroot}#{apparmor_profilesdir}/%{name}
install -m0644 distribution/systemd/* %{buildroot}%{_unitdir}

ln -s %{_sbindir}/service %{buildroot}%{_sbindir}/rc%{name}

%pre
#{apparmor_reload #{name}.service}
%service_add_pre %{name}.service

%post
%service_add_post %{name}.service

%preun
%service_del_preun %{name}.service

%postun
%service_del_postun %{name}.service

%files
%license COPYING
%doc README.md
%{_bindir}/%{name}
%{_sbindir}/rc%{name}
%{_datadir}/%{name}
#dir #{apparmor_profilesdir}
#config #{apparmor_profilesdir}/#{name}
%{_unitdir}/%{name}.service
%config %attr(0600,root,root) %{_sysconfdir}/%{name}.json

%changelog
