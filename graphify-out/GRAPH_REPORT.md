# Graph Report - workspace  (2026-09-29)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 3893 nodes · 11616 edges · 209 communities (145 shown, 64 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 365 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Community 0
- Community 1
- Community 2
- Community 3
- Community 4
- Community 5
- Community 6
- Community 7
- Community 8
- Community 9
- Community 10
- Community 11
- Community 12
- Community 13
- Community 14
- Community 15
- Community 16
- Community 17
- Community 18
- Community 19
- Community 20
- Community 21
- Community 22
- Community 23
- Community 24
- Community 25
- Community 26
- Community 27
- Community 28
- Community 29
- Community 30
- Community 31
- Community 32
- Community 33
- Community 34
- Community 35
- Community 36
- Community 37
- Community 38
- Community 39
- Community 40
- Community 41
- Community 42
- Community 43
- Community 44
- Community 45
- Community 46
- Community 47
- Community 48
- Community 49
- Community 50
- Community 51
- Community 52
- Community 53
- Community 54
- Community 55
- Community 56
- Community 57
- Community 58
- Community 59
- Community 60
- Community 61
- Community 62
- Community 63
- Community 64
- Community 65
- Community 66
- Community 67
- Community 68
- Community 69
- Community 70
- Community 71
- Community 72
- Community 73
- Community 74
- Community 75
- Community 76
- Community 77
- Community 78
- Community 79
- Community 80
- Community 81
- Community 82
- Community 83
- Community 84
- Community 85
- Community 86
- Community 87
- Community 88
- Community 89
- Community 90
- Community 91
- Community 92
- Community 93
- Community 94
- Community 95
- Community 96
- Community 97
- Community 98
- Community 99
- Community 100
- Community 101
- Community 102
- Community 103
- Community 104
- Community 105
- Community 106
- Community 107
- Community 108
- Community 109
- Community 110
- Community 111
- Community 112
- Community 113
- Community 114
- Community 115
- Community 116
- Community 117
- Community 118
- Community 119
- Community 120
- Community 121
- Community 122
- Community 123
- Community 124
- Community 125
- Community 126
- Community 127
- Community 128
- Community 129
- Community 130
- Community 131
- Community 132
- Community 133
- Community 134
- Community 135
- Community 136
- Community 137
- Community 138
- Community 139
- Community 140
- Community 141
- Community 142
- Community 143
- Community 144
- Community 145
- Community 146
- Community 147
- Community 148
- Community 149
- Community 150
- Community 151
- Community 152
- Community 153
- Community 154
- Community 155
- Community 156
- Community 157
- Community 158
- Community 159
- Community 160
- Community 161
- Community 162
- Community 163
- Community 164
- Community 165
- Community 166
- Community 167
- Community 168
- Community 169
- Community 170
- Community 171
- Community 172
- Community 173
- Community 174
- Community 175
- Community 176
- Community 177
- Community 178
- Community 179
- Community 180
- Community 181
- Community 182
- Community 183
- Community 184
- Community 185
- Community 186
- Community 187
- Community 188
- Community 189
- Community 190
- Community 192
- Community 193
- Community 194
- Community 195
- Community 196
- Community 197
- Community 201
- Community 202
- Community 203
- Community 204
- Community 208
- Community 210

## God Nodes (most connected - your core abstractions)
1. `SendErrorResponse()` - 206 edges
2. `SendOK()` - 147 edges
3. `SendJSONResponse()` - 122 edges
4. `PostPara()` - 122 edges
5. `E` - 105 edges
6. `constructor()` - 64 edges
7. `d` - 60 edges
8. `PostBool()` - 56 edges
9. `P` - 53 edges
10. `Logger` - 49 edges

## Surprising Connections (you probably didn't know these)
- `startupSequence()` --calls--> `initACME()`  [INFERRED]
  src/start.go → src/acme.go
- `initAPIs()` --calls--> `FSHandler()`  [INFERRED]
  src/api.go → src/router.go
- `main()` --calls--> `initAPIs()`  [INFERRED]
  src/main.go → src/api.go
- `UpdateProxyBasicAuthCredentials()` --calls--> `SaveReverseProxyConfig()`  [INFERRED]
  src/reverseproxy.go → src/config.go
- `main()` --calls--> `registerHealthEndpoints()`  [INFERRED]
  src/main.go → src/health.go

## Import Cycles
- None detected.

## Communities (209 total, 64 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (92): hostnameSuffixFilter, ZrFilter, net/http.Request, net/http.ResponseWriter, requestCanceler, handleListAccessRules(), handleListBlacklisted(), handleListQuickBan() (+84 more)

### Community 1 - "Community 1"
Cohesion: 0.01
Nodes (212): go_pkg_github_com_go_acme_lego_v5_challenge, go_pkg_github_com_go_acme_lego_v5_providers_dns_abion, go_pkg_github_com_go_acme_lego_v5_providers_dns_acmedns, go_pkg_github_com_go_acme_lego_v5_providers_dns_active24, go_pkg_github_com_go_acme_lego_v5_providers_dns_alidns, go_pkg_github_com_go_acme_lego_v5_providers_dns_aliesa, go_pkg_github_com_go_acme_lego_v5_providers_dns_allinkl, go_pkg_github_com_go_acme_lego_v5_providers_dns_alwaysdata (+204 more)

### Community 2 - "Community 2"
Cohesion: 0.02
Nodes (111): addDecoration(), _addLineToZone(), _addMouseDownListeners(), addOscHandler(), addRefreshCallback(), _applyScrollModifier(), _areCoordsInSelection(), areSelectionValuesReversed() (+103 more)

### Community 3 - "Community 3"
Cohesion: 0.05
Nodes (91): go_pkg_github_com_microcosm_cc_bluemonday, imuslab.com/zoraxy/mod/dynamicproxy.CaptchaConfig, handleAddTrustedProxy(), handleAttachRuleToHost(), handleBlacklistEnable(), handleBulkUpdateTrustedProxies(), handleCountryBlacklistAdd(), handleCountryBlacklistRemove() (+83 more)

### Community 4 - "Community 4"
Cohesion: 0.03
Nodes (11): addEscHandler(), compositionstart(), d, handleFocus(), hook(), P, register(), registerEscHandler() (+3 more)

### Community 6 - "Community 6"
Cohesion: 0.06
Nodes (28): net.Conn, net.Listener, net.UDPAddr, net.UDPConn, sync.WaitGroup, AccessRule, Controller, AccessRule (+20 more)

### Community 7 - "Community 7"
Cohesion: 0.12
Nodes (13): go_pkg_fmt, go_pkg_github_com_google_uuid, go_pkg_github_com_shirou_gopsutil_v4_cpu, go_pkg_github_com_shirou_gopsutil_v4_mem, go_pkg_math, go_pkg_os, go_pkg_os_signal, go_pkg_runtime (+5 more)

### Community 8 - "Community 8"
Cohesion: 0.05
Nodes (14): addLineToLink(), addMarker(), c(), clearAllMarkers(), clearMarkers(), createInstance(), _getEntryIdKey(), L() (+6 more)

### Community 9 - "Community 9"
Cohesion: 0.08
Nodes (15): _announceCharacters(), _convertViewportColToCharacterIndex(), getBufferElements(), getJoinedCharacters(), getLine(), _getWordAt(), h(), _isCharWordSeparator() (+7 more)

### Community 10 - "Community 10"
Cohesion: 0.06
Nodes (47): addEncoding(), addProtocol(), _clearLiveRegion(), clearRange(), clearTextureAtlas(), constructor(), _createAccessibilityTreeNode(), disable() (+39 more)

### Community 11 - "Community 11"
Cohesion: 0.11
Nodes (27): ConfigTemplate, go_pkg_archive_zip, go_pkg_compress_gzip, go_pkg_crypto_sha512, go_pkg_encoding_base64, go_pkg_encoding_json, go_pkg_github_com_go_webauthn_webauthn_protocol, go_pkg_github_com_go_webauthn_webauthn_webauthn (+19 more)

### Community 12 - "Community 12"
Cohesion: 0.15
Nodes (13): go_pkg_github_com_shirou_gopsutil_v4_disk, go_pkg_github_com_shirou_gopsutil_v4_net, go_pkg_golang_org_x_text_cases, go_pkg_golang_org_x_text_language, go_pkg_math_rand, go_pkg_regexp, dashboardBandwidthRecord, globalProxyTimeoutSettings (+5 more)

### Community 13 - "Community 13"
Cohesion: 0.10
Nodes (46): testing.T, Controller, newTestController(), TestAddTrustedProxy(), TestAddTrustedProxy_CIDR_RebuildCache(), TestGetClientIP_CFConnectingIP(), TestGetClientIP_NilParent_Fallback(), TestGetClientIP_TrustDisabled_UsesRemoteAddr() (+38 more)

### Community 14 - "Community 14"
Cohesion: 0.04
Nodes (45): Breaking Change, Bugfixupdate for big release of V3, read update notes from V3 if you are still on V2, IMPORTANT: V3 is a big rewrite and it is incompatible with V2! There is NO migration, if you want to stay on V2, please use V2 branch!, This release tidied up the contribution by [Teifun2](https://github.com/Teifun2) and added a new way to generate DNS challenge based certificate (e.g. wildcards) from Let's Encrypt without changing any environment variables. This also fixes a few previous ACME module EAB settings bug related to concurrent save., v2.6.1 May 31 2023, v2.6.2 Jun 4 2023, v2.6.3 Jun 8 2023, v2.6.4 Jun 15 2023 (+37 more)

### Community 15 - "Community 15"
Cohesion: 0.07
Nodes (16): AccessRuleCreatedEvent, BlacklistedIPBlockedEvent, BlacklistToggledEvent, CustomEvent, eventManager, Listener, TestListener, ListenerID (+8 more)

### Community 16 - "Community 16"
Cohesion: 0.07
Nodes (8): _addStyle(), _applyMinimumContrast(), createRow(), getColor(), _getContrastCache(), i(), setColor(), v()

### Community 17 - "Community 17"
Cohesion: 0.05
Nodes (41): access, acme, auth, Contributing, Core Modules, database, Dependencies, Development & Testing (+33 more)

### Community 18 - "Community 18"
Cohesion: 0.06
Nodes (12): _cancelCallback(), clear(), _getCorrectBufferLength(), n(), put(), r(), _removeIntersectingLinks(), _requestCallback() (+4 more)

### Community 19 - "Community 19"
Cohesion: 0.05
Nodes (35): Active/active control-plane migration, Migration safety, Required implementation order, Acceptance run, Backup and restore, Configure DRBD, Configure quorum and fencing, Install files (+27 more)

### Community 20 - "Community 20"
Cohesion: 0.05
Nodes (36): API Endpoints, CAPTCHA Exception Rules, CAPTCHA Gating Feature for Zoraxy, CAPTCHA not showing, Cloudflare Turnstile, Code Structure, Configuration, Configuration Storage (+28 more)

### Community 21 - "Community 21"
Cohesion: 0.09
Nodes (29): AccessRule, Router, ProxyEndpoint, basicAuthExceptionMatched(), BasicAuthExceptionRule, ProxyEndpoint, newBasicAuthEndpoint(), pathRule() (+21 more)

### Community 22 - "Community 22"
Cohesion: 0.11
Nodes (11): github.com/go-webauthn/webauthn/webauthn.Credential, PasskeyCredential, AuthRouter, PasskeyCredential, AuthRouter, normalizeUsername(), userToResponse(), GroupPolicy (+3 more)

### Community 24 - "Community 24"
Cohesion: 0.10
Nodes (19): github.com/gorilla/websocket.Conn, golang.org/x/crypto/ssh.Client, golang.org/x/crypto/ssh.Session, io.WriteCloser, NewReverseProxy(), singleJoiningSlash(), Manager, NewSSHProxyManager() (+11 more)

### Community 25 - "Community 25"
Cohesion: 0.12
Nodes (17): Config, time.Duration, TestHandlePing(), PingIP(), TCPPing(), UDPPing(), Monitor, Target (+9 more)

### Community 26 - "Community 26"
Cohesion: 0.08
Nodes (23): Options, RouterOption, GatewayServer, github.com/gorilla/sessions.CookieStore, net.IPNet, os/exec.Cmd, sync.RWMutex, time.Ticker (+15 more)

### Community 27 - "Community 27"
Cohesion: 0.14
Nodes (30): GetDriveStat(), printAndLog(), PrintSystemHardwareDebugMessage(), wmicGetinfo(), GetCPUArch(), GetCPUFreq(), GetCPUHardware(), GetCPUInfo() (+22 more)

### Community 28 - "Community 28"
Cohesion: 0.21
Nodes (18): CloudflareTurnstileResponse, GoogleRecaptchaResponse, TemplateData, BasicAuthUnhashedCredentials, go_pkg_bytes, go_pkg_crypto_tls, go_pkg_github_com_c0va23_go_proxyprotocol, go_pkg_github_com_gorilla_sessions (+10 more)

### Community 29 - "Community 29"
Cohesion: 0.11
Nodes (12): go_pkg_context, go_pkg_crypto_sha256, go_pkg_database_sql, go_pkg_github_com_gorilla_websocket, go_pkg_github_com_jackc_pgx_v5_stdlib, go_pkg_github_com_pires_go_proxyproto, go_pkg_golang_org_x_crypto_ssh, go_pkg_golang_org_x_net_http2 (+4 more)

### Community 30 - "Community 30"
Cohesion: 0.11
Nodes (17): time.Month, DataLoader, NewDataLoader(), generateDateRange(), mergeDailySummaryExports(), newBoundedCounters(), Collector, NewDailySummary() (+9 more)

### Community 31 - "Community 31"
Cohesion: 0.13
Nodes (20): PluginAPIKey, ManagerOptions, APIKeyManager, checkSupportHotRebuild(), Plugin, Manager, Manager, NewPluginManager() (+12 more)

### Community 32 - "Community 32"
Cohesion: 0.13
Nodes (6): github.com/gorilla/sessions.Options, AuthAgent, Hash(), NewAuthenticationAgent(), newSessionOptions(), UpdateProxyBasicAuthCredentials()

### Community 33 - "Community 33"
Cohesion: 0.14
Nodes (22): AuthRouterOptions, net/http.Client, AuthRouter, NewAuthRouter(), cleanSplit(), headerCookieRedact(), headerCopyAll(), headerCopyExcluded() (+14 more)

### Community 34 - "Community 34"
Cohesion: 0.09
Nodes (25): AuthenticationProvider, CaptchaConfig, HeaderRewriteRules, ProxyType, net/http.Server, net.PacketConn, RequestCountPerIpTable, RouterOption (+17 more)

### Community 35 - "Community 35"
Cohesion: 0.10
Nodes (22): Config, ExceptionRule, ExceptionType, Provider, AuthRouter, requestPathWithinPrefix(), TestRequestPathWithinPrefix(), CheckException() (+14 more)

### Community 36 - "Community 36"
Cohesion: 0.15
Nodes (4): crypto/tls.Config, github.com/quic-go/quic-go/http3.Server, Router, ProxyEndpoint

### Community 37 - "Community 37"
Cohesion: 0.10
Nodes (21): AuthMethod, HeaderDirection, HeaderRewriteOptions, GetDefaultPermissionPolicy(), PermissionsPolicy, InjectPermissionPolicyHeader(), TestInjectPermissionPolicyHeader(), CustomHeadersIncludeDynamicVariables() (+13 more)

### Community 38 - "Community 38"
Cohesion: 0.11
Nodes (10): io.Writer, log.Logger, net/http.Flusher, net/http.RoundTripper, maxLatencyWriter, writeFlusher, addXForwardedForHeader(), copyHeader() (+2 more)

### Community 39 - "Community 39"
Cohesion: 0.14
Nodes (9): DownloadablePlugin, PluginUpdateInfo, Manager, downloadableVersionIsNewer(), downloadFileTo(), Plugin, Manager, storeVersionIsNewer() (+1 more)

### Community 40 - "Community 40"
Cohesion: 0.10
Nodes (13): AutheliaRouter, AutheliaRouterOptions, AuthentikRouter, AuthentikRouterOptions, NewAutheliaRouter(), NewAuthentikRouter(), ResetAccount(), GetRecommendedBackendType() (+5 more)

### Community 41 - "Community 41"
Cohesion: 0.16
Nodes (14): AtomicActivator, Candidate, CandidateBuilder, CandidateRuntime, RetireErrorHandler, testCandidate, testCandidateBuilder, testCandidateRuntime (+6 more)

### Community 42 - "Community 42"
Cohesion: 0.15
Nodes (13): Handler, os.File, RotateOption, GetSessionKey(), NewAPIKeyManager(), UXOptimizer, NewDockerOptimizer(), InitEventSystem() (+5 more)

### Community 43 - "Community 43"
Cohesion: 0.14
Nodes (13): crypto/tls.Certificate, crypto/tls.ClientHelloInfo, crypto/x509.Certificate, ExportConfigAsZip(), Controller, Options, NewAccessController(), getCertPairs() (+5 more)

### Community 44 - "Community 44"
Cohesion: 0.22
Nodes (9): handlerStore, PostgresStore, Revision, Store, testActivator, testStore, context.Context, encoding/json.RawMessage (+1 more)

### Community 45 - "Community 45"
Cohesion: 0.10
Nodes (15): DpcoreOptions, crypto/x509.CertPool, net/http.Transport, joinURLPath(), newBaseTransport(), NewDynamicProxyCore(), newTestCert(), sharedServerName() (+7 more)

### Community 46 - "Community 46"
Cohesion: 0.15
Nodes (5): GetDefaultHeaderRewriteRules(), HeaderRewriteRules, ProxyEndpoint, VirtualDirectoryEndpoint, Upstream

### Community 47 - "Community 47"
Cohesion: 0.17
Nodes (14): ActorResolver, ControlPlane, HTTPHandler, ManagementRouter, Validator, Store, NewControlPlane(), TestControlPlaneRegistersRevisionAndNodeStatusAPIs() (+6 more)

### Community 48 - "Community 48"
Cohesion: 0.18
Nodes (12): ErrorTemplateType, ProxyHandler, Router, ProxyEndpoint, VirtualDirectoryEndpoint, serveProxyRequestError(), GetHeaderVariableValuesFromRequest(), PopulateRequestHeaderVariables() (+4 more)

### Community 49 - "Community 49"
Cohesion: 0.16
Nodes (20): calculateCPUUsage(), getCPUStats(), GetCPUUsageUsingProcStat(), StartBackgroundMonitor(), GetCPUUsage(), GetNumericRAMUsage(), GetRAMUsage(), TestCalculateCPUUsage() (+12 more)

### Community 50 - "Community 50"
Cohesion: 0.14
Nodes (18): CaDef, EABConfig, go_pkg_crypto, go_pkg_crypto_ecdsa, go_pkg_crypto_elliptic, go_pkg_crypto_rand, go_pkg_crypto_x509, go_pkg_crypto_x509_pkix (+10 more)

### Community 51 - "Community 51"
Cohesion: 0.14
Nodes (5): WhitelistEntry, deepCopy(), AccessRule, Controller, AccessRule

### Community 52 - "Community 52"
Cohesion: 0.14
Nodes (12): CountryInfo, geoDataSet, StoreOptions, sync/atomic.Pointer, Store, NewGeoDb(), TestResolveCountryCodeFromIP(), RequestInfo (+4 more)

### Community 53 - "Community 53"
Cohesion: 0.26
Nodes (4): getDefaultBranding(), AuthRouter, writeBrandingAsset(), BrandingConfig

### Community 54 - "Community 54"
Cohesion: 0.17
Nodes (17): buildOriginalRequestURL(), buildSSOAuthRedirect(), ensureAbsoluteHTTPURL(), hasAbsoluteHTTPScheme(), hostnameOnly(), requestScheme(), ssoRedirectWouldLoop(), TestBuildLoginRedirectURL_HostOnlyConfig() (+9 more)

### Community 55 - "Community 55"
Cohesion: 0.16
Nodes (12): os.FileMode, getRandomPort(), initACME(), parseACMEFileMode(), ACMEHandler, IsPortInUse(), NewACME(), IsSupportedCA() (+4 more)

### Community 56 - "Community 56"
Cohesion: 0.11
Nodes (22): AuthEndpoints, go_pkg_encoding_csv, go_pkg_errors, go_pkg_flag, go_pkg_github_com_armon_go_radix, go_pkg_github_com_gorilla_csrf, go_pkg_github_com_jellydator_ttlcache_v3, go_pkg_github_com_syndtr_goleveldb_leveldb (+14 more)

### Community 57 - "Community 57"
Cohesion: 0.15
Nodes (17): net/http.Header, RequireWebsocketHeaderCopy(), IsProxmox(), copyHeader(), addXForwardedForHeader(), injectUserDefinedHeaders(), removeHeaders(), rewriteUserAgent() (+9 more)

### Community 58 - "Community 58"
Cohesion: 0.21
Nodes (8): go_pkg_bufio, go_pkg_github_com_likexian_whois, go_pkg_golang_org_x_net_icmp, go_pkg_golang_org_x_net_ipv4, go_pkg_golang_org_x_net_webdav, go_pkg_net, OpenedPort, ScanPorts()

### Community 59 - "Community 59"
Cohesion: 0.11
Nodes (28): go_pkg_encoding_binary, go_pkg_encoding_xml, go_pkg_image, go_pkg_image_color, go_pkg_unicode, encoding/xml.Attr, defaultAssetContent(), checkAspectRatio() (+20 more)

### Community 60 - "Community 60"
Cohesion: 0.20
Nodes (17): sync.Map, boundedIncr(), evictLeastFrequent(), mapLen(), TestBoundedIncrConcurrent(), TestBoundedIncrCountsExistingKeys(), TestBoundedIncrTrimsLowFrequencyEntries(), TestNewIncrFnDispatch() (+9 more)

### Community 61 - "Community 61"
Cohesion: 0.11
Nodes (19): ProxyType, runUpdateRoutineWithVersion(), CopyFile(), copyFile(), UpdateFrom307To308(), convertV314ToV315(), UpdateFrom314To315(), BasicAuthCredentials (+11 more)

### Community 62 - "Community 62"
Cohesion: 0.21
Nodes (18): RouterOption, initAPIs(), RegisterAccessRuleAPIs(), RegisterACMEAndAutoRenewerAPIs(), RegisterAuthenticationHandlerAPIs(), RegisterHTTPProxyAPIs(), RegisterMDNSAPIs(), RegisterNetworkUtilsAPIs() (+10 more)

### Community 63 - "Community 63"
Cohesion: 0.20
Nodes (9): Updater, UpdaterStatus, sync/atomic.Bool, sync/atomic.Int64, sync/atomic.Value, Store, Store, jitteredInterval() (+1 more)

### Community 64 - "Community 64"
Cohesion: 0.18
Nodes (4): AuthRouter, isAjaxRequest(), AuthRouter, getDefaultOptions()

### Community 65 - "Community 65"
Cohesion: 0.16
Nodes (13): CertificateInfoJSON, ExpiredCerts, ACMEHandler, contains(), LoadCertInfoJSON(), TestExtractIssuerNameFromPEM(), NewAutoRenewer(), CertExpireSoon() (+5 more)

### Community 66 - "Community 66"
Cohesion: 0.29
Nodes (14): PluginMiddlewareOptions, net/http.HandlerFunc, PluginAuthMiddleware, NewPluginAuthMiddleware(), initRestAPI(), RegisterAccessRuleRestAPI(), RegisterHTTPProxyRestAPI(), RegisterMDNSRestAPI() (+6 more)

### Community 67 - "Community 67"
Cohesion: 0.27
Nodes (3): option, singleton, New()

### Community 68 - "Community 68"
Cohesion: 0.24
Nodes (15): allChecksPass(), currentReadinessChecks(), handleClusterStatus(), handleLiveness(), nodeRole(), readinessHandler(), registerHealthEndpoints(), TestClusterStatusExposesConvergenceFields() (+7 more)

### Community 69 - "Community 69"
Cohesion: 0.15
Nodes (14): _batchedMemoryCleanup(), _createElement(), fillViewportRows(), fire(), getBlankLine(), getNullCell(), _reflow(), _reflowLarger() (+6 more)

### Community 71 - "Community 71"
Cohesion: 0.17
Nodes (8): net/http.Handler, NewPluginFileSystemUIRouter(), PathRouter, NewPathRouter(), RewriteURL(), FSHandler(), isHTMLFilePath(), PluginUiDebugRouter

### Community 72 - "Community 72"
Cohesion: 0.16
Nodes (15): ArOZInfo, Server, filterGrepResults(), NewInfoServer(), TestFilterGrepResults(), TestGetArOZInfoHandler(), TestGetArOZInfoHandlerWithIcon(), TestGetCPUInfoHandler() (+7 more)

### Community 73 - "Community 73"
Cohesion: 0.12
Nodes (16): acmeDeregisterSpecialRoutingRule(), acmeRegisterSpecialRoutingRule(), restartACMEHandler(), main(), SetupCloseHandler(), parseCSV(), downloadAndValidateGeoDBCsv(), DownloadGeoDBUpdate() (+8 more)

### Community 74 - "Community 74"
Cohesion: 0.18
Nodes (14): Database, newDatabase(), NewBoltDatabase(), TestNewBoltDatabase(), TestNewTable(), TestTableExists(), TestSnapshotProducesReadableConsistentDatabase(), NewDB() (+6 more)

### Community 76 - "Community 76"
Cohesion: 0.15
Nodes (15): github.com/go-webauthn/webauthn/webauthn.SessionData, time.Time, Organization, WhoisIpLookupEntry, WHOISResult, parseDate(), ParseWhoisIpData(), ParseWHOISResponse() (+7 more)

### Community 78 - "Community 78"
Cohesion: 0.16
Nodes (15): Router, newTestRateLimitRouter(), newTestRequest(), TestHandleRateLimit_BareHeaderIPStillCounts(), TestHandleRateLimit_CountsPerRealClient(), TestHandleRateLimit_DirectConnection(), TestHandleRateLimit_SpoofedHeaderFromUntrustedPeerIgnored(), Router (+7 more)

### Community 80 - "Community 80"
Cohesion: 0.13
Nodes (14): Basic HTTP/1.1 Request, Configuration, Expected Output, Features, Force HTTP/1.1, How It Works, HTTP/1.1 Test Server, Notes (+6 more)

### Community 81 - "Community 81"
Cohesion: 0.19
Nodes (11): Activator, database/sql.DB, Follow(), FollowNode(), Store, TestFollowKeepsServingAfterRejectedRevision(), TestFollowNodeReportsDesiredRejectedAndAppliedRevisions(), TestFollowNodeStopsWhenConvergenceCannotBeReported() (+3 more)

### Community 82 - "Community 82"
Cohesion: 0.26
Nodes (10): clientConnector, configuration, defaultHandler, Filter, hostnameFilter, monitor, serverConnector, newHandler() (+2 more)

### Community 83 - "Community 83"
Cohesion: 0.26
Nodes (6): defaultProxy, proxy, Socket, newProxy(), tryCloseRead(), tryCloseWrite()

### Community 84 - "Community 84"
Cohesion: 0.24
Nodes (8): ipRange, ipRangeV6, trie, net.IP, constrctTrieTree(), ipv4ToUint64(), ipv6ToUint64Pair(), newTrie()

### Community 85 - "Community 85"
Cohesion: 0.23
Nodes (4): AuthRouter, GatewayServer, NewGatewayServer(), validateTOTPCode()

### Community 86 - "Community 86"
Cohesion: 0.20
Nodes (6): a(), addCsiHandler(), addDcsHandler(), registerCsiHandler(), registerDcsHandler(), t()

### Community 87 - "Community 87"
Cohesion: 0.22
Nodes (6): ACMEUser, crypto/ecdsa.PrivateKey, crypto.Signer, github.com/go-acme/lego/v5/acme.ExtendedAccount, accountDBKey(), ACMEHandler

### Community 88 - "Community 88"
Cohesion: 0.24
Nodes (3): fileExists(), Database, isDirectory()

### Community 89 - "Community 89"
Cohesion: 0.24
Nodes (8): io.Reader, ResponseRewriteRuleSet, ReverseProxy, TestReplaceLocationHost(), TestReplaceLocationHostRelative(), ReplaceLocationHost(), replaceLocationHost(), IsValidTLSFile()

### Community 90 - "Community 90"
Cohesion: 0.17
Nodes (6): Plugin, DecodeForwardRequestPayload(), EncodeForwardRequestPayload(), PathRouter, DynamicSniffForwardRequest, SniffHandler

### Community 91 - "Community 91"
Cohesion: 0.35
Nodes (13): TestSaveAfterMidnightKeepsPreviousDay(), NewStatisticCollector(), summaryKeyOf(), BenchmarkRecordRequest(), clearDatabase(), getNewDatabase(), TestGetCurrentRealtimeStatIntervalId(), TestLoadSummaryOfDay() (+5 more)

### Community 92 - "Community 92"
Cohesion: 0.31
Nodes (11): convertV307ToV308(), v307BasicAuthCredentials, v307BasicAuthExceptionRule, v307BasicAuthUnhashedCredentials, v307HeaderDirection, v307PermissionsPolicy, v307ProxyEndpoint, v307UserDefinedHeader (+3 more)

### Community 94 - "Community 94"
Cohesion: 0.24
Nodes (7): CheckSession(), generateSessionID(), SessionStore, HandleVerification(), NewSessionStore(), VerifyCloudflareToken(), VerifyGoogleRecaptchaToken()

### Community 95 - "Community 95"
Cohesion: 0.35
Nodes (11): Field, ProviderInfo, extractConfigStruct(), extractInternalConfigStruct(), fileExists(), getExcludedDNSProviders(), getExcludedDNSProvidersNT61(), isExcludedDNSProvider() (+3 more)

### Community 96 - "Community 96"
Cohesion: 0.26
Nodes (4): controlPlaneStore, NodeStatus, statusHandlerStore, testNodeStatusStore

### Community 97 - "Community 97"
Cohesion: 0.26
Nodes (5): testManagementRouter, embed.FS, net/http.ServeMux, NewPluginEmbedUIRouter(), PluginUiRouter

### Community 98 - "Community 98"
Cohesion: 0.20
Nodes (3): DB, github.com/syndtr/goleveldb/leveldb.Batch, github.com/syndtr/goleveldb/leveldb.DB

### Community 99 - "Community 99"
Cohesion: 0.17
Nodes (11): Building, Docker Compose, Docker Run, Environment, Extra Hosts, Plugins, Ports, Usage (+3 more)

### Community 100 - "Community 100"
Cohesion: 0.23
Nodes (5): ProxyEndpoint, Router, ProxyEndpoint, StringInArray(), StringInArrayIgnoreCase()

### Community 101 - "Community 101"
Cohesion: 0.29
Nodes (5): ProxyHandler, ProxyEndpoint, handleAuthProviderRouting(), handleBasicAuth(), ProxyHandler

### Community 103 - "Community 103"
Cohesion: 0.27
Nodes (5): golang.org/x/net/webdav.Handler, NewWebDAVServer(), NewWebServer(), WebDAVServer, WebServerOptions

### Community 104 - "Community 104"
Cohesion: 0.32
Nodes (4): BlockingPath, Options, Handler, NewPathRuleHandler()

### Community 105 - "Community 105"
Cohesion: 0.33
Nodes (4): RedirectRule, RuleTable, NewRuleTable(), ReplaceSpecialCharacters()

### Community 106 - "Community 106"
Cohesion: 0.31
Nodes (3): TrustedProxy, Controller, loadDefaultTrustedProxiesFromCSV()

### Community 108 - "Community 108"
Cohesion: 0.24
Nodes (4): RequestCountPerIpTable, ProxyHandler, Router, ProxyEndpoint

### Community 109 - "Community 109"
Cohesion: 0.25
Nodes (9): reservedIPNode, reservedIPRadixTree, testing.B, getReservedIPZone(), initReservedIPTrees(), BenchmarkGetReservedIPZone(), BenchmarkGetReservedIPZone_PrivateIPv4(), BenchmarkGetReservedIPZone_PrivateIPv6() (+1 more)

### Community 110 - "Community 110"
Cohesion: 0.25
Nodes (9): github.com/gorilla/websocket.Dialer, github.com/gorilla/websocket.Upgrader, net/url.URL, Options, NewProxy(), ProxyHandler(), TestProxy(), Options (+1 more)

### Community 112 - "Community 112"
Cohesion: 0.22
Nodes (3): AuthRouter, AuthRouter, NewAuthRouter()

### Community 113 - "Community 113"
Cohesion: 0.25
Nodes (4): _addCallbacks(), dispose(), n, provideLinks()

### Community 116 - "Community 116"
Cohesion: 0.31
Nodes (10): bufio.Reader, net/http/httptest.Server, dialUpgrade(), newUpgradeProxy(), newUpgradeUpstream(), TestNonUpgradeRequestIsUnaffected(), TestUpgradeGateAllowsPlainRequest(), TestUpgradeIsBlockedByDefault() (+2 more)

### Community 117 - "Community 117"
Cohesion: 0.33
Nodes (4): github.com/shirou/gopsutil/v4/net.IOCountersStat, dashboardBandwidthTracker, nicCounter, initDashboardBandwidthTracker()

### Community 118 - "Community 118"
Cohesion: 0.33
Nodes (4): LogFile, Viewer, ViewerOption, NewLogViewer()

### Community 121 - "Community 121"
Cohesion: 0.28
Nodes (6): NodeStatusHandler, nodeStatusResponse, NodeStatusStore, NewNodeStatusHandler(), TestNodeStatusHandlerListsConvergenceState(), TestNodeStatusHandlerReportsUnavailableStore()

### Community 122 - "Community 122"
Cohesion: 0.31
Nodes (5): defaultServerConnector, initializer, loggingInitializer, newServerConnector(), newLoggingInitializer()

### Community 123 - "Community 123"
Cohesion: 0.26
Nodes (12): cleanup(), getenv(), main(), popen(), run(), start_zerotier(), start_zoraxy(), os (+4 more)

### Community 124 - "Community 124"
Cohesion: 0.31
Nodes (5): net/http.Response, H2CRoundTripper, ReverseProxy, copyHeader(), copyResponse()

### Community 125 - "Community 125"
Cohesion: 0.36
Nodes (5): getRandomUpstreamByWeight(), RouteManager, Upstream, calculateStdDev(), TestRandomUpstreamSelection()

### Community 127 - "Community 127"
Cohesion: 0.64
Nodes (8): assertTotals(), newTestBandwidthTracker(), nicReading(), TestBandwidthBaselinesNewInterface(), TestBandwidthFirstSampleOnlyBaselines(), TestBandwidthIgnoresCounterReset(), TestBandwidthResetsOnlyAtLocalMidnight(), TestBandwidthSurvivesDisappearingInterface()

### Community 129 - "Community 129"
Cohesion: 0.36
Nodes (5): Router, ProxyEndpoint, defaultHostSpecificTlsBehavior(), GetDefaultHostSpecificTlsBehavior(), HostSpecificTlsBehavior

### Community 130 - "Community 130"
Cohesion: 0.43
Nodes (3): ExploitsRequestResponseType, Detector, NewExploitDetector()

### Community 131 - "Community 131"
Cohesion: 0.32
Nodes (4): Store, ipv4ToUInt32(), isIPv4InRange(), isIPv6InRange()

### Community 132 - "Community 132"
Cohesion: 0.36
Nodes (3): github.com/go-webauthn/webauthn/webauthn.WebAuthn, GatewayServer, newWebAuthnFromRequest()

### Community 133 - "Community 133"
Cohesion: 0.32
Nodes (6): github.com/grandcat/zeroconf.Server, net.Interface, MDNSHost, NetworkHost, getMacAddr(), NewMDNS()

### Community 134 - "Community 134"
Cohesion: 0.43
Nodes (5): golang.org/x/oauth2.Config, OAuth2RouterOptions, OAuth2Router, NewOAuth2Router(), ttlcache.Cache

### Community 135 - "Community 135"
Cohesion: 0.39
Nodes (7): LogConfig, Logger, HandleUpdateLogConfig(), LoadLogConfig(), SaveLogConfig(), SizeStringToBytes(), TestSizeStringToBytes()

### Community 137 - "Community 137"
Cohesion: 0.25
Nodes (8): AuthenticationProvider, AuthProvider, BasicAuthCredentials, BasicAuthExceptionRule, HeaderRewriteRules, ProxyType, v315ProxyEndpoint, VirtualDirectoryEndpoint

### Community 138 - "Community 138"
Cohesion: 0.29
Nodes (5): proxyProtocolInitializer, net.Addr, formatHeader(), newProxyProtocolInitializer(), parseAddress()

### Community 139 - "Community 139"
Cohesion: 0.38
Nodes (7): AuthenticationProvider, AuthExceptionType, AuthMethod, ZorxAuthExceptionRule, BasicAuthCredentials, BasicAuthExceptionRule, Record

### Community 140 - "Community 140"
Cohesion: 0.43
Nodes (6): LogicalDisk, parseDfOutput(), TestParseDfOutputBSDLayout(), TestParseDfOutputGNULayout(), TestParseDfOutputMalformedLines(), TestParseDfOutputMixedGoodAndBad()

### Community 142 - "Community 142"
Cohesion: 0.29
Nodes (10): go_pkg_crypto_rsa, crypto/rsa.PrivateKey, crypto/rsa.PublicKey, BytesToPrivateKey(), BytesToPublicKey(), DecryptWithPrivateKey(), EncryptWithPublicKey(), GenerateKeyPair() (+2 more)

### Community 143 - "Community 143"
Cohesion: 0.35
Nodes (6): go_pkg_github_com_go_ping_ping, DiscoveredHost, inc(), ScanCIDRRange(), ScanIpRange(), sortByIP()

### Community 145 - "Community 145"
Cohesion: 0.40
Nodes (3): defaultDialer, logger, newDialer()

### Community 146 - "Community 146"
Cohesion: 0.40
Nodes (3): Dialer, routingDialer, newRoutingDialer()

### Community 147 - "Community 147"
Cohesion: 0.33
Nodes (3): maxLatencyWriter, sync.Mutex, time.Timer

### Community 150 - "Community 150"
Cohesion: 0.40
Nodes (4): certificate_revisions, config_revisions, controller_leases, node_status

### Community 151 - "Community 151"
Cohesion: 0.50
Nodes (4): BulkVdirAction, ClassifyBulkVdir(), VirtualDirectoryEndpoint, TestClassifyBulkVdir()

### Community 152 - "Community 152"
Cohesion: 0.50
Nodes (5): github.com/go-acme/lego/v5/challenge.Provider, GetDnsChallengeProviderByName(), GetDNSProviderByJsonConfig(), TestACMEDNSConfigStructureReflector(), GetProviderConfigStructure()

### Community 155 - "Community 155"
Cohesion: 0.40
Nodes (5): IsLoopbackIPOrDomain(), IsSSHConnectable(), IsWebSSHSupported(), HandleCreateProxySession(), HandleWebSshSupportCheck()

### Community 157 - "Community 157"
Cohesion: 0.60
Nodes (4): IsValidMacAddress(), sendPacket(), WakeTarget(), magicPacket

### Community 158 - "Community 158"
Cohesion: 0.40
Nodes (4): Build for Windows 7 (NT6.1), DNS Challenge Update Data Structure Code Generator, Module Usage, Usage

### Community 161 - "Community 161"
Cohesion: 0.83
Nodes (4): io.ReadWriteCloser, copyWithDeadlineRefresh(), setConnDeadline(), tunnelUpgradedConnection()

### Community 162 - "Community 162"
Cohesion: 0.67
Nodes (3): ContentSecurityPolicy, GetDefaultContentSecurityPolicy(), InjectContentSecurityPolicyHeader()

### Community 164 - "Community 164"
Cohesion: 0.67
Nodes (3): ExpectedRevision(), TestExpectedRevision(), TestExpectedRevisionRejectsMissingAndWildcard()

### Community 165 - "Community 165"
Cohesion: 0.50
Nodes (3): Example, Install, WebsocketProxy [![GoDoc](https://godoc.org/github.com/koding/websocketproxy?status.svg)](https://godoc.org/github.com/koding/websocketproxy) [![Build Status](https://travis-ci.org/koding/websocketproxy.svg)](https://travis-ci.org/koding/websocketproxy)

### Community 166 - "Community 166"
Cohesion: 0.50
Nodes (4): countActiveProxyConnections(), countSyncMapEntries(), getBandwidthToday(), HandleDashboardOverview()

### Community 174 - "Community 174"
Cohesion: 0.67
Nodes (3): archive/zip.Writer, addFileToZip(), ZipFiles()

### Community 176 - "Community 176"
Cohesion: 0.67
Nodes (3): DomainIsSelfSigned(), DomainUsesTLS(), HandleCheckSiteSupportTLS()

### Community 177 - "Community 177"
Cohesion: 0.67
Nodes (3): TestHandleTraceRoute(), TraceRoute(), traceroute()

### Community 208 - "Community 208"
Cohesion: 0.50
Nodes (3): Sender, go_pkg_net_smtp, NewEmailSender()

### Community 210 - "Community 210"
Cohesion: 0.50
Nodes (4): go_pkg_golang_org_x_net_http2_h2c, NewH2CRoundTripper(), TestH2CRoundTripper_CheckServerSupportsH2C(), TestH2CRoundTripper_RoundTrip()

## Knowledge Gaps
- **209 isolated node(s):** `config_revisions`, `node_status`, `certificate_revisions`, `controller_leases`, `imuslab.com/zoraxy` (+204 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 915 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **64 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `SendErrorResponse()` connect `Community 3` to `Community 0`, `Community 32`, `Community 64`, `Community 132`, `Community 39`, `Community 167`, `Community 136`, `Community 135`, `Community 11`, `Community 155`, `Community 176`, `Community 53`, `Community 27`, `Community 93`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **Why does `NewDynamicProxyCore()` connect `Community 45` to `Community 100`, `Community 110`, `Community 15`, `Community 116`, `Community 89`, `Community 90`, `Community 28`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Why does `AuthAgent` connect `Community 32` to `Community 0`, `Community 103`, `Community 40`, `Community 42`, `Community 56`, `Community 26`, `Community 62`?**
  _High betweenness centrality (0.009) - this node is a cross-community bridge._
- **What connects `config_revisions`, `node_status`, `certificate_revisions` to the rest of the system?**
  _209 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.02969902332070959 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.009389671361502348 - nodes in this community are weakly interconnected._
- **Should `Community 2` be split into smaller, more focused modules?**
  _Cohesion score 0.019107847137222904 - nodes in this community are weakly interconnected._