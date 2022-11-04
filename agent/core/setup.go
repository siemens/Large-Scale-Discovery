/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"fmt"
	"github.com/siemens/GoScans/banner"
	"github.com/siemens/GoScans/discovery"
	"github.com/siemens/GoScans/nfs"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssh"
	"github.com/siemens/GoScans/ssl"
	"github.com/siemens/GoScans/webcrawler"
	"github.com/siemens/GoScans/webenum"
	"large-scale-discovery/agent/config"
	"large-scale-discovery/log"
	"time"
)

// Setup executes the setup functions of required scan modules in order to prepare the system for scanning.
func Setup() (error, string) {

	// Get tagged logger
	logger := log.GetLogger().Tagged("Setup")

	// Get config
	conf := config.GetConfig()

	// Setup Banner
	errBanner := banner.Setup(logger)
	if errBanner != nil {
		return errBanner, banner.Label
	}

	// Setup Discovery
	errNmap := discovery.Setup(logger, conf.Paths.NmapDir, conf.Paths.Nmap)
	if errNmap != nil {
		return errNmap, discovery.Label
	}

	// Setup Nfs
	errNfs := nfs.Setup(logger)
	if errNfs != nil {
		return errNfs, nfs.Label
	}

	// Setup Smb
	errSmb := smb.Setup(logger)
	if errSmb != nil {
		return errSmb, smb.Label
	}

	// Setup Ssh
	errSsh := ssh.Setup(logger)
	if errSsh != nil {
		return errSsh, ssh.Label
	}

	// Setup Ssl
	errSsl := ssl.Setup(logger)
	if errSsl != nil {
		return errSsl, ssl.Label
	}

	// Setup Webcrawler
	errWebcrawler := webcrawler.Setup(logger)
	if errWebcrawler != nil {
		return errWebcrawler, webcrawler.Label
	}

	// Setup Webenum
	errWebenum := webenum.Setup(logger)
	if errWebenum != nil {
		return errWebenum, webenum.Label
	}

	// Return nil as everything went fine
	return nil, ""
}

// CheckSetup tests whether the setup functions of requires scan modules were executed successfully.
func CheckSetup() (error, string) {

	// Get config
	conf := config.GetConfig()

	// Run Banner setup test
	errBanner := banner.CheckSetup()
	if errBanner != nil {
		return errBanner, banner.Label
	}

	// Run Nfs setup test
	errNfs := nfs.CheckSetup()
	if errNfs != nil {
		return errNfs, nfs.Label
	}

	// Run Smb setup test
	errSmb := smb.CheckSetup()
	if errSmb != nil {
		return errSmb, smb.Label
	}

	// Run Discovery setup test
	errNmap := discovery.CheckSetup(conf.Paths.NmapDir, conf.Paths.Nmap)
	if errNmap != nil {
		return errNmap, discovery.Label
	}

	// Run Ssh setup test
	errSsh := ssh.CheckSetup()
	if errSsh != nil {
		return errSsh, ssh.Label
	}

	// Run Ssl setup test
	errSsl := ssl.CheckSetup()
	if errSsl != nil {
		return errSsl, ssl.Label
	}

	// Run Webcrawler setup test
	errWebcrawler := webcrawler.CheckSetup()
	if errWebcrawler != nil {
		return errWebcrawler, webcrawler.Label
	}

	// Run Webenum setup test
	errWebenum := webenum.CheckSetup()
	if errWebenum != nil {
		return errWebenum, webenum.Label
	}

	// Return nil as everything went fine
	return nil, ""
}

// CheckConfig tests current configuration values by trying to initialize scan modules with them. This allows to
// discover invalid configurations at startup, instead of during runtime. Dynamic target arguments are replaced by
// dummy data.
func CheckConfig() error {

	// Dummy scan target arguments
	dummyLogger := log.GetLogger().Tagged("CheckConfig")
	dummyTarget := "127.0.0.1"
	dummyPort := 0
	dummyOtherNames := []string{"a", "b"}
	dummyNetworkTimeout := time.Second * 0
	dummyHttpUserAgent := "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:60.0) Gecko/20100101 Firefox/60.0"

	// Get config
	conf := config.GetConfig()

	// Run Banner test
	_, errBanner := banner.NewScanner(dummyLogger, dummyTarget, dummyPort, "tcp", dummyNetworkTimeout, dummyNetworkTimeout)
	if errBanner != nil {
		return fmt.Errorf("'%s': %s", banner.Label, errBanner)
	}

	// Run Discovery test
	_, errNmap := discovery.NewScanner(
		dummyLogger,
		[]string{dummyTarget},
		conf.Paths.Nmap,
		[]string{
			"-PE",
			"-PP",
			"-PS21,22,25,23,80,111,179,443,445,1433,1521,3189,3306,3389,5800,5900,8000,8008,8080,8443",
			"-PA80,21000",
			"-sS",
			"-O",
			"-p0-65535",
			"-sV",
			"-T4",
			"--randomize-hosts",
			"--host-timeout",
			"6h",
			"--max-retries",
			"2",
			"--script",
			"address-info,afp-serverinfo,ajp-auth,ajp-methods,amqp-info,auth-owners,backorifice-info,bitcoinrpc-info,cassandra-info,clock-skew,creds-summary,dns-nsid,dns-recursion,dns-service-discovery,epmd-info,finger,flume-master-info,freelancer-info,ftp-anon,ftp-bounce,ganglia-info,giop-info,gopher-ls,hadoop-datanode-info,hadoop-jobtracker-info,hadoop-namenode-info,hadoop-secondary-namenode-info,hadoop-tasktracker-info,hbase-master-info,hbase-region-info,hddtemp-info,hnap-info,Http-auth,Http-cisco-anyconnect,Http-cors,Http-generator,Http-git,Http-open-proxy,Http-robots.txt,Http-svn-enum,Http-webdav-scan,ike-version,imap-capabilities,imap-ntlm-info,ip-https-discover,ipv6-node-info,irc-info,iscsi-info,jdwp-info,knx-gateway-info,maxdb-info,mongodb-databases,mongodb-info,ms-sql-info,ms-sql-ntlm-info,mysql-info,nat-pmp-info,nbstat,ncp-serverinfo,netbus-info,nntp-ntlm-info,openlookup-info,pop3-capabilities,pop3-ntlm-info,quake1-info,quake3-info,quake3-master-getservers,realvnc-auth-bypass,rmi-dumpregistry,rpcinfo,rtsp-methods,servicetags,sip-methods,smb-security-mode,smb-protocols,smtp-commands,smtp-ntlm-info,snmp-hh3c-logins,snmp-info,snmp-interfaces,snmp-netstat,snmp-processes,snmp-sysdescr,snmp-win32-services,snmp-win32-shares,snmp-win32-software,snmp-win32-users,socks-auth-info,socks-open-proxy,ssh-hostkey,sshv1,ssl-known-key,sstp-discover,telnet-ntlm-info,tls-nextprotoneg,upnp-info,ventrilo-info,vnc-info,wdb-version,weblogic-t3-info,wsdd-discover,x11-access,xmlrpc-methods,xmpp-info,vnc-title,acarsd-info,afp-showmount,ajp-headers,ajp-request,allseeingeye-info,bitcoin-getaddr,bitcoin-info,citrix-enum-apps,citrix-enum-servers-xml,citrix-enum-servers,coap-resources,couchdb-databases,couchdb-stats,daytime,db2-das-info,dict-info,drda-info,duplicates,gpsd-info,Http-affiliate-id,Http-apache-negotiation,Http-apache-server-status,Http-cross-domain-policy,Http-frontpage-login,Http-gitweb-projects-enum,Http-php-version,Http-qnap-nas-info,Http-vlcstreamer-ls,Http-vuln-cve2010-0738,Http-vmware-path-vuln,Http-vuln-cve2011-3192,Http-vuln-cve2014-2126,Http-vuln-cve2014-2127,Http-vuln-cve2014-2128,ip-forwarding,ipmi-cipher-zero,ipmi-version,membase-Http-info,memcached-info,mqtt-subscribe,msrpc-enum,ncp-enum-users,netbus-auth-bypass,nfs-ls,nfs-showmount,nfs-statfs,omp2-enum-targets,oracle-tns-version,rdp-enum-encryption,redis-info,rfc868-time,riak-Http-info,rsync-list-modules,rusers,smb-mbenum,ssh2-enum-algos,stun-info,telnet-encryption,tn3270-screen,versant-info,voldemort-info,vuze-dht-info,xdmcp-discover,supermicro-ipmi-conf,cccam-version,docker-version,enip-info,fox-info,iax2-version,jdwp-version,netbus-version,pcworx-info,s7-info,teamspeak2-version",
			"--script",
			"vmware-version,tls-ticketbleed,smb2-time,smb2-security-mode,smb2-capabilities,smb-vuln-ms17-010,smb-double-pulsar-backdoor,openwebnet-discovery,Http-vuln-cve2017-1001000,Http-security-headers,Http-cookie-flags,ftp-syst,cics-info",
		},
		true,
		[]string{instanceIp}, // Exclude local IP from scans, scan would have extended privileges discovering content that isn't visible from the outside
		conf.Modules.Discovery.BlacklistFile,
		[]string{".local", "sub1.local", "sub2.local", "third-party.com"},
		conf.Modules.Discovery.LdapServer,
		conf.Authentication.Ldap.Domain,
		conf.Authentication.Ldap.User,
		conf.Authentication.Ldap.Password,
		dummyNetworkTimeout,
	)
	if errNmap != nil {
		return fmt.Errorf("'%s': %s", discovery.Label, errNmap)
	}

	// Run NFS test
	_, errNfs := nfs.NewScanner(
		dummyLogger,
		dummyTarget,
		3,
		3,
		[]string{"a", "b", "c"},
		[]string{"a"},
		[]string{".a"},
		time.Date(2008, 01, 01, 00, 00, 00, 00, time.UTC),
		-1,
		true,
		time.Second*5,
	)
	if errNfs != nil {
		return fmt.Errorf("'%s': %s", nfs.Label, errNfs)
	}

	// Run Ssh test
	_, errSsh := ssh.NewScanner(dummyLogger, dummyTarget, dummyPort, dummyNetworkTimeout)
	if errSsh != nil {
		return fmt.Errorf("'%s': %s", ssh.Label, errSsh)
	}

	// Prepare os-specific trust store for Ssl module
	if len(conf.Modules.Ssl.CustomTruststoreFile) == 0 {
		errGenOsTruststore := generateTruststoreOs(SslOsTruststoreFile)
		if errGenOsTruststore != nil {
			return fmt.Errorf("'%s': %s", ssl.Label, errGenOsTruststore)
		}
	}

	// Run Webcrawler test
	_, errWebcrawler := webcrawler.NewScanner(
		dummyLogger,
		dummyTarget,
		dummyPort,
		dummyOtherNames,
		true,
		3,
		3,
		true,
		true,
		conf.Modules.Webcrawler.Download,
		conf.Modules.Webcrawler.DownloadPath,
		conf.Authentication.Webcrawler.Domain,
		conf.Authentication.Webcrawler.User,
		conf.Authentication.Webcrawler.Password,
		dummyHttpUserAgent,
		"",
		dummyNetworkTimeout,
	)
	if errWebcrawler != nil {
		return fmt.Errorf("'%s': %s", webcrawler.Label, errWebcrawler)
	}

	// Run Webenum test
	_, errWebenum := webenum.NewScanner(
		dummyLogger,
		dummyTarget,
		dummyPort,
		dummyOtherNames,
		true,
		conf.Authentication.Webenum.Domain,
		conf.Authentication.Webenum.User,
		conf.Authentication.Webenum.Password,
		WebenumProbesFile,
		true,
		dummyHttpUserAgent,
		"",
		dummyNetworkTimeout,
	)
	if errWebenum != nil {
		return fmt.Errorf("'%s': %s", webenum.Label, errWebenum)
	}

	// Run OS specific tests
	errSpecific := checkConfigDependant()
	if errSpecific != nil {
		return errSpecific
	}

	// Return nil as everything went fine
	return nil
}
