# Large-Scale Discovery

Large-Scale Discovery is a solution for discovering assets, services and contents and metadata within large IT/OT
networks. It relies on a few basic well-known open-source tools ([Nmap](https://nmap.org/)
, [SSLyze](https://github.com/nabla-c0d3/sslyze/releases))
and is extended with additional scan logic and several custom scan modules.

Results are collected in a structured repository of things visible and accessible. Data can be searched, filtered,
combined, post-processed and visualized as desired via a common database browsers, custom scripts or third party BI
software.

- Discover or identify assets and contents unknown before
- Discover implicit IT security issues by analyzing result data
- Get insight into an invisible situation

Related services or solutions are [Shodan](https://www.shodan.io/), respectively, [IVRE](https://ivre.rocks/)
and [Natlas](https://natlas.io/). In contrast, Large-Scale Discovery is going deeper into common communication
protocols (HTTP, SMB,...) to discover contents and includes user and access management.

## Key Features

- Web interface for scan configuration and control
- Multiple scan scopes, groups and users possible
- Continuous scans to keep data updated
- Scan agents can be put into dedicated network zones
- Randomization of scan targets for load distribution
- Restriction of scan times to local working hours (e.g. during office times or outside)
- Scalable configuration and parallelization
- Discovery of potential hostnames from different sources
- Data expansion with information from remote repositories, e.g., Active Directory
- Versatile scan module configuration
- Lowest possible dependency stack
- ...

## Scan modules

- Asset and service discovery
- Banner grabbing
- SMB crawling
- NFS crawling
- Web crawling
- Web enumeration
- SSH enumeration
- SSL enumeration
- Nuclei vulnerability enumeration

## Components

The solution comprises several software components. Some of them are optional, and they can be hosted individually and
in parallel to scale:

- *Scope manager* (`manager`): (Sensitive) central component managing existing scan scopes, respective storage
  databases, views, access rights and credentials.
- *Broker* (`broker`): Intermediary component, linking scan agents to assigned scan scopes, deciding and distributing
  scan tasks across connected scan agents.
- *Scan agent* (`agent`): Executing component containing scan logic and taking care of the execution of actual network
  scans, as instructed by the broker.


- *Web Interface*  (`web_backend` + `web_frontend`): Control component allowing users to configure scan scopes, scan
  settings, manage result access rights and to monitor scan progress. Alternatively, you may configure scan scopes
  manually or write your own scripts.
- *Importer* (`importer`): Optional utility component to sync scan scope inputs with external sources. This component is
  not necessary for _custom_ scan scopes, which are directly populated from the user input via the web interface. The
  importer can be deployed to import scan inputs from external sources, e.g. a network or asset repository.
- *PgProxy* (`pgproxy`): Optional PostgreSQL proxy component for secure, controlled access to scan scope databases.

  <img src="Readme_components.png" alt="Components and Interfaces" />

# Code Philosophy

This project intends to keep the initial effort to get rolling as little as possible. Things should work out of the box.
Just click 'compile', and you will notice only a few dependencies that need to be prepared manually (
see [Preparation](#preparation)). If you compile with build tag 'prod', all development settings and code will be
hardened, respectively, stripped.

# Development

## GIT configuration

If you plan to submit code, please set up your GIT, e.g.:

```
git config --global user.name "Name Surname"
git config --global user.email "...@domain.tld"
```

## Manager, broker & agent for scanning

#### Preparation

* Install a local PostgreSQL with the following configuration for **development**:

```
        Host: localhost     (do NOT enable listening on the external interface!)
        Port: 5432
        User: postgres      (administrative user required to manage databases)
        Pass: test123!$LSD  (do NOT use in production!)
```

* Edit `./_test/settings.go`and replace the dummy e-mail address.
* To avoid accidental commits, move `./_test/settings.go` to a custom/new Changelist (in GoLand, otherwise use a similar
  technique) as GIT is not designed to ignore tracked files.
* Download and install Nmap (https://nmap.org/) and SSLyze (https://github.com/nabla-c0d3/sslyze) to `./_bin/tools`. You
  can install them anywhere, but then you'll need to update the paths in `agent.conf` later on.
* **Configure the working directory** in your build configurations to `./_bin/` (we've prepared some necessary files
  there).

#### First launch

* Launch _manager_, it will generate a default `./_bin/manager.conf` file.
* Launch _broker_, it will generate a default `./_bin/broker.conf` file.
* Launch _agent_, it will generate a default `./_bin/agent.conf` file.
* Launch _agent_ **again** with `-setup` argument to prepare the operating system.
* Launch _manager_ **again**, it will prepare a development scan scope and some dummy data
* Launch _broker_  **again**, it will connect to the scope database (PostgreSQL) on localhost and initialize necessary
  database tables.
* Insert sample scan targets from `./_bin/t_discovery_sample.sql` (or better, your own) into the scan scope's respective
  input table `t_discovery`.

All components are pre-configured and should directly start interacting with each other after launching in development
mode. Similarly, in production mode, but with hardened configuration.

#### Scan process

* Launch _manager_, it holds available databases and knows/manages existing scan scopes (tables, views, access
  rights,...). It will create/migrate database tables if necessary. It provides an RPC API for other components to query
  and alter scopes and attributes.
* Launch _broker_, it will query known scopes, and their database connection details from the _manager_ and wait for _
  agent_ connections. It provides an RPC API for scan agents to request scan tasks and send scan results.
* Launch _agent_, it will connect to the _broker_ and request scan targets.
    * It will tell the _broker_ which tasks are currently running. The _broker_ will return scan tasks based on the
      associated scope's configuration.
* The _agent_ tells the _broker_ how many scan jobs of which kind are currently running. The _broker_ will decide which
  and how many more scan tasks shall be started additionally.
    * Main scan targets are taken from the scan scope's database table `t_discovery`. The `scan_started` attribute is
      set to indicate an active process.
    * Main scan results will be written back to the database tables `t_discovery_hosts`, `t_discovery_services`
      and `t_discovery_scripts`.
    * Submodule scan targets are extracted from the main scan results and kept in the _broker_'s cache (persistent
      SQLite DB).
    * Submodule scan targets are taken from the _broker_'s cache. The `scan_started` attribute is set to indicate an
      active process.
    * Submodule scan results will be written back to the respective scan scope database table, e.g. `t_ssl` and
      associated ones (`t_ssl_certificates`, `t_ssl_ciphers`,...).
    * You can open the cache `./_bin/broker.sqlite` with an SQLite browser and watch submodule targets being created and
      getting processed.
* Scan scopes are managed via the _web frontend_ (see below), which is connected to the _web backend_, which is
  connected to the _manager_.

## Web frontend for managing scan scopes

The frontend is a Vue 3, TypeScript, Vite, and PrimeVue single-page application. The web backend serves its
generated `web_frontend/dist/` files; it does not serve frontend source files.

### Prerequisites

- Node.js 20 LTS or later. CI currently uses Node.js 25.
- A running web backend for API requests.

### Development

Install the locked dependencies and rebuild the bundle when files change:

```bash
cd web_frontend
npm ci
npm run build:watch
```

In a second terminal, start the backend. It serves the generated `dist/` bundle and the REST API:

```bash
cd web_frontend
go run ../web_backend
```

For Vite hot-module replacement during frontend work, start the backend first and then run `npm run dev` in
`web_frontend`. The Vite server listens on `http://127.0.0.1:5173` and proxies `/api/v1` to the local backend.

### Production bundle

Run `npm run build` in `web_frontend`. Vite writes the optimized, content-hashed assets to `web_frontend/dist/`,
which the production backend serves.

### Frontend layout

- `src/main.ts` initializes Vue, PrimeVue, routing, and session restoration.
- `src/pages` contains route-level views; `src/components` contains feature and reusable UI components.
- `src/stores` owns client and server state; `src/api` owns HTTP, authentication, and normalized API errors.
- `tests/unit` contains Vitest unit and component tests; `tests/e2e` contains Playwright browser tests.

See [web_frontend/README.md](web_frontend/README.md) for the full frontend quality gates, architecture, and
Content Security Policy details.

# Production

## Compile

#### Building on Windows for Windows

Go to `./_bin` and execute `build_prod_win.bat` to create production builds. Production builds will:

- exclude development code (by applying build tag 'prod')
- apply more secure/realistic production settings
- strip debug symbols

#### Building on Linux for Linux

Go to `./_bin` and execute `build_prod_unix.sh` to create production builds. Production builds will:

- exclude development code (by applying build tag 'prod')
- apply more secure/realistic production settings
- strip debug symbols

It's easiest to not bother cross-compiling and just use a Linux machine (e.g. VM) to build for Linux.

## Deployment

You can deploy the single components on a single machine or install them on dedicated machines.
It is strongly recommended to deploy the `manager` component on a dedicated isolated machine as it holds some sensitive
access credentials.

Before the system is ready to run, you'll also need to:

- Generate production keys / certificates and put them into the respective component's `keys/` folder.
    - *manager.key* -> manager **ONLY!**
    - *manager.crt* -> manager + broker
    - *broker.key*  -> broker **ONLY!**
    - *broker.crt*  -> broker + agent
    - *backend.key* -> backend **ONLY!**
    - *backend.crt* -> backend **ONLY!**


- Configure the storage database. Open `manager.sqlite` and create an entry in the `db_servers` table.
  <img src="Readme_configuration_databases.png" alt="Database Configuration" />

# Screenshots

<kbd>
<img src="Readme_dashboard.png" alt="Dashboard" />
</kbd>
Dashboard
<br/><br/>
<kbd>
<img src="Readme_configuration_scopes.png" alt="Scopes Configuration" />
</kbd>
Scope configuration
<br/><br/>
<kbd>
<img src="Readme_configuration_views.png" alt="Views Configuration" />
</kbd>
Views configuration and access rights
<br/><br/>
<kbd>
<img src="Readme_configuration_scope.png" alt="Scan Settings" />
</kbd>
Scan settings

# Some entertainment

[![Development History](Readme_visualization_code.gif)](https://youtu.be/I6YeXmt6cGY)
<br />[[Full Video](https://youtu.be/I6YeXmt6cGY)]
