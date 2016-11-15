Name:       gobak
Version:    0.2
Release:    1
Summary:    A program for make increment backup for firebird
Packager:   Gordienko R. formeo@pahrm-tmn.ru
Vendor:     AO Pharmacy,Russia,Tyumen
URL:        http://www.pharm-tmn.ru/gobak
Source0:    gobak
Source1:    config.json   
Group:      Utilities/System  
License:    GPL

%description
Increment backup on golang

%description -l ru
Данный сервис работает в качестве демона, но возможен запуск и консольным приложением

%files
/opt/gobackup/gobak
/opt/gobackup/config.json


%install
rm -rf $RPM_BUILD_ROOT

install -d  $RPM_BUILD_ROOT/opt/gobackup

cp -f %{SOURCE0} $RPM_BUILD_ROOT/opt/gobackup
cp -n  %{SOURCE1} $RPM_BUILD_ROOT/opt/gobackup

%clean
rm -rf $RPM_BUILD_ROOT

%post
/opt/gobackup/gobak service install
echo "Configure config.json"

%preun
/opt/gobackup/gobak service uninstall
