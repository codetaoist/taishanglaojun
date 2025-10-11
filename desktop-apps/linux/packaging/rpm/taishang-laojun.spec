Name:           taishang-laojun
Version:        1.0.0
Release:        1%{?dist}
Summary:        TaishangLaojun Desktop Application

License:        MIT
URL:            https://github.com/taishanglaojun/desktop-apps
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  cmake >= 3.16
BuildRequires:  ninja-build
BuildRequires:  gcc
BuildRequires:  pkg-config
BuildRequires:  gtk4-devel >= 4.6
BuildRequires:  glib2-devel >= 2.66
BuildRequires:  json-c-devel >= 0.13
BuildRequires:  openssl-devel >= 1.1
BuildRequires:  sqlite-devel >= 3.31
BuildRequires:  libcurl-devel >= 7.68
BuildRequires:  libnotify-devel >= 0.7
BuildRequires:  libadwaita-devel >= 1.0
BuildRequires:  webkit2gtk4.1-devel >= 2.36
BuildRequires:  libappindicator-gtk3-devel >= 0.4
BuildRequires:  desktop-file-utils
BuildRequires:  libappstream-glib

Requires:       gtk4 >= 4.6
Requires:       glib2 >= 2.66
Requires:       json-c >= 0.13
Requires:       openssl-libs >= 1.1
Requires:       sqlite >= 3.31
Requires:       libcurl >= 7.68
Requires:       libnotify >= 0.7
Requires:       libadwaita >= 1.0
Requires:       webkit2gtk4.1 >= 2.36
Requires:       libappindicator-gtk3 >= 0.4

Recommends:     gstreamer1-plugins-good
Recommends:     gstreamer1-plugins-bad
Recommends:     gstreamer1-libav

%description
TaishangLaojun is a modern desktop application that provides an intuitive
interface for AI-powered conversations and productivity tools. Built with
GTK4 and modern C, it offers a native Linux experience with excellent
performance and integration.

Key features include:
- Modern GTK4-based user interface
- AI-powered conversation capabilities
- Cross-platform compatibility
- Extensible plugin system
- Comprehensive configuration management
- Multi-language support
- System tray integration
- Desktop notifications

%prep
%autosetup

%build
%cmake -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX=%{_prefix} \
    -DCMAKE_INSTALL_LIBDIR=%{_libdir} \
    -DCMAKE_INSTALL_BINDIR=%{_bindir} \
    -DCMAKE_INSTALL_DATADIR=%{_datadir} \
    -DCMAKE_INSTALL_MANDIR=%{_mandir} \
    -DENABLE_TESTS=ON \
    -DENABLE_DOCS=ON
%cmake_build

%install
%cmake_install

# Install desktop file
desktop-file-install \
    --dir=%{buildroot}%{_datadir}/applications \
    %{buildroot}%{_datadir}/applications/%{name}.desktop

# Install AppStream metadata
install -Dm644 resources/appstream/%{name}.appdata.xml \
    %{buildroot}%{_datadir}/metainfo/%{name}.appdata.xml

# Install manual page
install -Dm644 resources/man/%{name}.1 \
    %{buildroot}%{_mandir}/man1/%{name}.1

%check
%ctest

# Validate desktop file
desktop-file-validate %{buildroot}%{_datadir}/applications/%{name}.desktop

# Validate AppStream metadata
appstream-util validate-relax %{buildroot}%{_datadir}/metainfo/%{name}.appdata.xml

%post
%{_bindir}/update-desktop-database %{_datadir}/applications &> /dev/null || :
%{_bindir}/gtk-update-icon-cache %{_datadir}/icons/hicolor &> /dev/null || :

%postun
%{_bindir}/update-desktop-database %{_datadir}/applications &> /dev/null || :
%{_bindir}/gtk-update-icon-cache %{_datadir}/icons/hicolor &> /dev/null || :

%files
%license LICENSE
%doc README.md docs/BUILD.md
%{_bindir}/%{name}
%{_datadir}/applications/%{name}.desktop
%{_datadir}/metainfo/%{name}.appdata.xml
%{_datadir}/icons/hicolor/scalable/apps/%{name}.svg
%{_mandir}/man1/%{name}.1*

%changelog
* Mon Jan 06 2025 TaishangLaojun Team <team@taishanglaojun.com> - 1.0.0-1
- Initial release of TaishangLaojun Desktop Application
- Modern GTK4-based user interface implementation
- Core application framework with GObject-based architecture
- Configuration management system with JSON backend
- Comprehensive utility functions for system integration
- Desktop integration with .desktop file and AppStream metadata
- System tray integration and desktop notifications
- Multi-language support infrastructure
- Plugin system foundation
- Comprehensive test suite
- Complete build and packaging system for multiple formats
- Documentation and user manual
- CI/CD pipeline configuration

* Mon Dec 30 2024 TaishangLaojun Team <team@taishanglaojun.com> - 0.9.0-1
- Pre-release version for testing
- Basic application structure
- Initial GTK4 integration
- Core configuration system
- Basic utility functions