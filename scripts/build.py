#!/usr/bin/env python
#
# This is the Telegraf build script.
#
# Current caveats:
#   - Does not checkout the correct commit/branch (for now, you will need to do so manually)
#   - Has external dependencies for packaging (fpm) and uploading (boto)
#

import sys
import os
import subprocess
import time
import datetime
import shutil
import tempfile
import hashlib
import re

debug = False

# PACKAGING VARIABLES
INSTALL_ROOT_DIR = "/usr/bin"
LOG_DIR = "/var/log/telegraf"
SCRIPT_DIR = "/usr/lib/telegraf/scripts"
CONFIG_DIR = "/etc/telegraf"
LOGROTATE_DIR = "/etc/logrotate.d"

INIT_SCRIPT = "scripts/init.sh"
SYSTEMD_SCRIPT = "scripts/telegraf.service"
LOGROTATE_SCRIPT = "etc/logrotate.d/telegraf"
DEFAULT_CONFIG = "etc/telegraf.conf"
DEFAULT_WINDOWS_CONFIG = "etc/telegraf_windows.conf"
POSTINST_SCRIPT = "scripts/post-install.sh"
PREINST_SCRIPT = "scripts/pre-install.sh"

# META-PACKAGE VARIABLES
PACKAGE_LICENSE = "MIT"
PACKAGE_URL = "https://github.com/influxdata/telegraf"
MAINTAINER = "support@influxdb.com"
VENDOR = "InfluxData"
DESCRIPTION = "Plugin-driven server agent for reporting metrics into InfluxDB."

# SCRIPT START
prereqs = [ 'git', 'go' ]
optional_prereqs = [ 'fpm', 'rpmbuild' ]

fpm_common_args = "-f -s dir --log error \
 --vendor {} \
 --url {} \
 --license {} \
 --maintainer {} \
 --config-files {} \
 --config-files {} \
 --after-install {} \
 --before-install {} \
 --description \"{}\"".format(
    VENDOR,
    PACKAGE_URL,
    PACKAGE_LICENSE,
    MAINTAINER,
    CONFIG_DIR + '/telegraf.conf',
    LOGROTATE_DIR + '/telegraf',
    POSTINST_SCRIPT,
    PREINST_SCRIPT,
    DESCRIPTION)

targets = {
    'telegraf' : './cmd/telegraf/telegraf.go',
}

supported_builds = {
    'darwin': [ "amd64", "i386" ],
    'windows': [ "amd64", "i386" ],
    'linux': [ "amd64", "i386", "arm" ],
    'freebsd': [ "amd64" ]
}
supported_packages = {
    "darwin": [ "tar", "zip" ],
    "linux": [ "deb", "rpm", "tar", "zip" ],
    "windows": [ "zip" ],
    'freebsd': [ "tar" ]
}
supported_tags = {
    # "linux": {
    #     "amd64": ["sensors"]
    # }
}
prereq_cmds = {
    # "linux": "sudo apt-get install lm-sensors libsensors4-dev"
}

def run(command, allow_failure=False, shell=False):
    out = None
    if debug:
        print("[DEBUG] {}".format(command))
    try:
        if shell:
            out = subprocess.check_output(command, stderr=subprocess.STDOUT, shell=shell)
        else:
            out = subprocess.check_output(command.split(), stderr=subprocess.STDOUT)
        out = out.decode("utf8")
    except subprocess.CalledProcessError as e:
        print("")
        print("")
        print("Executed command failed!")
        print("-- Command run was: {}".format(command))
        print("-- Failure was: {}".format(e.output))
        if allow_failure:
            print("Continuing...")
            return None
        else:
            print("")
            print("Stopping.")
            sys.exit(1)
    except OSError as e:
        print("")
        print("")
        print("Invalid command!")
        print("-- Command run was: {}".format(command))
        print("-- Failure was: {}".format(e))
        if allow_failure:
            print("Continuing...")
            return out
        else:
            print("")
            print("Stopping.")
            sys.exit(1)
    else:
        return out

def create_temp_dir(prefix=None):
    if prefix is None:
        return tempfile.mkdtemp(prefix="telegraf-build.")
    else:
        return tempfile.mkdtemp(prefix=prefix)

def get_current_version():
    command = "git describe --always --tags --abbrev=0"
    out = run(command)
    return out.strip()

def get_current_commit(short=False):
    command = None
    if short:
        command = "git log --pretty=format:'%h' -n 1"
    else:
        command = "git rev-parse HEAD"
    out = run(command)
    return out.strip('\'\n\r ')

def get_current_branch():
    command = "git rev-parse --abbrev-ref HEAD"
    out = run(command)
    return out.strip()

def get_system_arch():
    arch = os.uname()[4]
    if arch == "x86_64":
        arch = "amd64"
    return arch

def get_system_platform():
    if sys.platform.startswith("linux"):
        return "linux"
    else:
        return sys.platform

def get_go_version():
    out = run("go version")
    matches = re.search('go version go(\S+)', out)
    if matches is not None:
        return matches.groups()[0].strip()
    return None

def check_path_for(b):
    def is_exe(fpath):
        return os.path.isfile(fpath) and os.access(fpath, os.X_OK)

    for path in os.environ["PATH"].split(os.pathsep):
        path = path.strip('"')
        full_path = os.path.join(path, b)
        if os.path.isfile(full_path) and os.access(full_path, os.X_OK):
            return full_path

def check_environ(build_dir = None):
    print("\nChecking environment:")
    for v in [ "GOPATH", "GOBIN", "GOROOT" ]:
        print("\t- {} -> {}".format(v, os.environ.get(v)))

    cwd = os.getcwd()
    if build_dir == None and os.environ.get("GOPATH") and os.environ.get("GOPATH") not in cwd:
        print("\n!! WARNING: Your current directory is not under your GOPATH. This may lead to build failures.")

def check_prereqs():
    print("\nChecking for dependencies:")
    for req in prereqs:
        path = check_path_for(req)
        if path is None:
            path = '?'
        print("\t- {} -> {}".format(req, path))
    for req in optional_prereqs:
        path = check_path_for(req)
        if path is None:
            path = '?'
        print("\t- {} (optional) -> {}".format(req, path))
    print("")

def upload_packages(packages, bucket_name=None, nightly=False):
    if debug:
        print("[DEBUG] upload_packags: {}".format(packages))
    try:
        import boto
        from boto.s3.key import Key
    except ImportError:
        print "!! Cannot upload packages without the 'boto' python library."
        return 1
    print("Uploading packages to S3...")
    print("")
    c = boto.connect_s3()
    if bucket_name is None:
        bucket_name = 'get.influxdb.org/telegraf'
    bucket = c.get_bucket(bucket_name.split('/')[0])
    print("\t - Using bucket: {}".format(bucket_name))
    for p in packages:
        if '/' in bucket_name:
            # Allow for nested paths within the bucket name (ex:
            # bucket/telegraf). Assuming forward-slashes as path
            # delimiter.
            name = os.path.join('/'.join(bucket_name.split('/')[1:]),
                                os.path.basename(p))
        else:
            name = os.path.basename(p)
        if bucket.get_key(name) is None or nightly:
            print("\t - Uploading {} to {}...".format(name, bucket_name))
            k = Key(bucket)
            k.key = name
            if nightly:
                n = k.set_contents_from_filename(p, replace=True)
            else:
                n = k.set_contents_from_filename(p, replace=False)
            k.make_public()
        else:
            print("\t - Not uploading {}, already exists.".format(p))
    print("")

def build(version=None,
          branch=None,
          commit=None,
          platform=None,
          arch=None,
          nightly=False,
          rc=None,
          race=False,
          clean=False,
          outdir=".",
          goarm_version="6"):
    print("-------------------------")
    print("")
    print("Build plan:")
    print("\t- version: {}".format(version))
    if rc:
        print("\t- release candidate: {}".format(rc))
    print("\t- commit: {}".format(commit))
    print("\t- branch: {}".format(branch))
    print("\t- platform: {}".format(platform))
    print("\t- arch: {}".format(arch))
    if arch == 'arm' and goarm_version:
        print("\t- ARM version: {}".format(goarm_version))
    print("\t- nightly? {}".format(str(nightly).lower()))
    print("\t- race enabled? {}".format(str(race).lower()))
    print("")

    if not os.path.exists(outdir):
        os.makedirs(outdir)
    elif clean and outdir != '/':
        print("Cleaning build directory...")
        shutil.rmtree(outdir)
        os.makedirs(outdir)

    if rc:
        # If a release candidate, update the version information accordingly
        version = "{}rc{}".format(version, rc)

    # Set architecture to something that Go expects
    if arch == 'i386':
        arch = '386'
    elif arch == 'x86_64':
        arch = 'amd64'

    print("Starting build...")
    for b, c in targets.items():
        if platform == 'windows':
            b = b + '.exe'
        print("\t- Building '{}'...".format(os.path.join(outdir, b)))
        build_command = ""
        build_command += "GOOS={} GOARCH={} ".format(platform, arch)
        if arch == "arm" and goarm_version:
            if goarm_version not in ["5", "6", "7", "arm64"]:
                print("!! Invalid ARM build version: {}".format(goarm_version))
            build_command += "GOARM={} ".format(goarm_version)
        build_command += "go build -o {} ".format(os.path.join(outdir, b))
        if race:
            build_command += "-race "
        if platform in supported_tags:
            if arch in supported_tags[platform]:
                build_tags = supported_tags[platform][arch]
                for build_tag in build_tags:
                    build_command += "-tags "+build_tag+" "
        go_version = get_go_version()
        if "1.4" in go_version:
            build_command += "-ldflags=\"-X main.buildTime '{}' ".format(datetime.datetime.utcnow().isoformat())
            build_command += "-X main.Version {} ".format(version)
            build_command += "-X main.Branch {} ".format(get_current_branch())
            build_command += "-X main.Commit {}\" ".format(get_current_commit())
        else:
            build_command += "-ldflags=\"-X main.buildTime='{}' ".format(datetime.datetime.utcnow().isoformat())
            build_command += "-X main.Version={} ".format(version)
            build_command += "-X main.Branch={} ".format(get_current_branch())
            build_command += "-X main.Commit={}\" ".format(get_current_commit())
        build_command += c
        run(build_command, shell=True)
    print("")

def create_dir(path):
    try:
        os.makedirs(path)
    except OSError as e:
        print(e)

def rename_file(fr, to):
    try:
        os.rename(fr, to)
    except OSError as e:
        print(e)
        # Return the original filename
        return fr
    else:
        # Return the new filename
        return to

def copy_file(fr, to):
    try:
        shutil.copy(fr, to)
    except OSError as e:
        print(e)

def create_package_fs(build_root):
    print("\t- Creating a filesystem hierarchy from directory: {}".format(build_root))
    # Using [1:] for the path names due to them being absolute
    # (will overwrite previous paths, per 'os.path.join' documentation)
    dirs = [ INSTALL_ROOT_DIR[1:], LOG_DIR[1:], SCRIPT_DIR[1:], CONFIG_DIR[1:], LOGROTATE_DIR[1:] ]
    for d in dirs:
        create_dir(os.path.join(build_root, d))
        os.chmod(os.path.join(build_root, d), 0o755)

def package_scripts(build_root, windows=False):
    print("\t- Copying scripts and sample configuration to build directory")
    if windows:
        shutil.copyfile(DEFAULT_WINDOWS_CONFIG, os.path.join(build_root, "telegraf.conf"))
        os.chmod(os.path.join(build_root, "telegraf.conf"), 0o644)
    else:
        shutil.copyfile(INIT_SCRIPT, os.path.join(build_root, SCRIPT_DIR[1:], INIT_SCRIPT.split('/')[1]))
        os.chmod(os.path.join(build_root, SCRIPT_DIR[1:], INIT_SCRIPT.split('/')[1]), 0o644)
        shutil.copyfile(SYSTEMD_SCRIPT, os.path.join(build_root, SCRIPT_DIR[1:], SYSTEMD_SCRIPT.split('/')[1]))
        os.chmod(os.path.join(build_root, SCRIPT_DIR[1:], SYSTEMD_SCRIPT.split('/')[1]), 0o644)
        shutil.copyfile(LOGROTATE_SCRIPT, os.path.join(build_root, LOGROTATE_DIR[1:], "telegraf"))
        os.chmod(os.path.join(build_root, LOGROTATE_DIR[1:], "telegraf"), 0o644)
        shutil.copyfile(DEFAULT_CONFIG, os.path.join(build_root, CONFIG_DIR[1:], "telegraf.conf"))
        os.chmod(os.path.join(build_root, CONFIG_DIR[1:], "telegraf.conf"), 0o644)

def go_get():
    print("Retrieving Go dependencies...")
    run("go get github.com/sparrc/gdm")
    run("gdm restore -f Godeps_windows")
    run("gdm restore")

def generate_md5_from_file(path):
    m = hashlib.md5()
    with open(path, 'rb') as f:
        while True:
            data = f.read(4096)
            if not data:
                break
            m.update(data)
    return m.hexdigest()

def build_packages(build_output, version, pkg_arch, nightly=False, rc=None, iteration=1):
    outfiles = []
    tmp_build_dir = create_temp_dir()
    if debug:
        print("[DEBUG] build_output = {}".format(build_output))
    try:
        print("-------------------------")
        print("")
        print("Packaging...")
        for p in build_output:
            # Create top-level folder displaying which platform (linux, etc)
            create_dir(os.path.join(tmp_build_dir, p))
            for a in build_output[p]:
                current_location = build_output[p][a]
                # Create second-level directory displaying the architecture (amd64, etc)p
                build_root = os.path.join(tmp_build_dir, p, a)
                # Create directory tree to mimic file system of package
                create_dir(build_root)
                if p == 'windows':
                    package_scripts(build_root, windows=True)
                else:
                    create_package_fs(build_root)
                    # Copy in packaging and miscellaneous scripts
                    package_scripts(build_root)
                # Copy newly-built binaries to packaging directory
                for b in targets:
                    if p == 'windows':
                        b = b + '.exe'
                        to = os.path.join(build_root, b)
                    else:
                        to = os.path.join(build_root, INSTALL_ROOT_DIR[1:], b)
                    fr = os.path.join(current_location, b)
                    print("\t- [{}][{}] - Moving from '{}' to '{}'".format(p, a, fr, to))
                    copy_file(fr, to)
                # Package the directory structure
                for package_type in supported_packages[p]:
                    print("\t- Packaging directory '{}' as '{}'...".format(build_root, package_type))
                    name = "telegraf"
                    # Reset version, iteration, and current location on each run
                    # since they may be modified below.
                    package_version = version
                    package_iteration = iteration
                    current_location = build_output[p][a]

                    if package_type in ['zip', 'tar']:
                        if nightly:
                            name = '{}-nightly_{}_{}'.format(name, p, a)
                        else:
                            name = '{}-{}-{}_{}_{}'.format(name, package_version, package_iteration, p, a)
                    if package_type == 'tar':
                        # Add `tar.gz` to path to reduce package size
                        current_location = os.path.join(current_location, name + '.tar.gz')
                    if rc is not None:
                        package_iteration = "0.rc{}".format(rc)
                    saved_a = a
                    if pkg_arch is not None:
                        a = pkg_arch
                    if a == '386':
                        a = 'i386'
                    if package_type == 'zip':
                        zip_command = "cd {} && zip {}.zip ./*".format(
                            build_root,
                            name)
                        run(zip_command, shell=True)
                        run("mv {}.zip {}".format(os.path.join(build_root, name), current_location), shell=True)
                        outfile = os.path.join(current_location, name+".zip")
                        outfiles.append(outfile)
                        print("\t\tMD5 = {}".format(generate_md5_from_file(outfile)))
                    else:
                        fpm_command = "fpm {} --name {} -a {} -t {} --version {} --iteration {} -C {} -p {} ".format(
                            fpm_common_args,
                            name,
                            a,
                            package_type,
                            package_version,
                            package_iteration,
                            build_root,
                            current_location)
                        if pkg_arch is not None:
                            a = saved_a
                        if package_type == "rpm":
                            fpm_command += "--depends coreutils "
                            fpm_command += "--depends lsof"
                        out = run(fpm_command, shell=True)
                        matches = re.search(':path=>"(.*)"', out)
                        outfile = None
                        if matches is not None:
                            outfile = matches.groups()[0]
                        if outfile is None:
                            print("[ COULD NOT DETERMINE OUTPUT ]")
                        else:
                            # Strip nightly version (the unix epoch) from filename
                            if nightly and package_type in ['deb', 'rpm']:
                                outfile = rename_file(outfile, outfile.replace("{}-{}".format(version, iteration), "nightly"))
                            outfiles.append(os.path.join(os.getcwd(), outfile))
                            # Display MD5 hash for generated package
                            print("\t\tMD5 = {}".format(generate_md5_from_file(outfile)))
        print("")
        if debug:
            print("[DEBUG] package outfiles: {}".format(outfiles))
        return outfiles
    finally:
        # Cleanup
        shutil.rmtree(tmp_build_dir)

def print_usage():
    print("Usage: ./build.py [options]")
    print("")
    print("Options:")
    print("\t --outdir=<path> \n\t\t- Send build output to a specified path. Defaults to ./build.")
    print("\t --arch=<arch> \n\t\t- Build for specified architecture. Acceptable values: x86_64|amd64, 386, arm, or all")
    print("\t --goarm=<arm version> \n\t\t- Build for specified ARM version (when building for ARM). Default value is: 6")
    print("\t --platform=<platform> \n\t\t- Build for specified platform. Acceptable values: linux, windows, darwin, or all")
    print("\t --version=<version> \n\t\t- Version information to apply to build metadata. If not specified, will be pulled from repo tag.")
    print("\t --pkgarch=<package-arch> \n\t\t- Package architecture if different from <arch>")
    print("\t --commit=<commit> \n\t\t- Use specific commit for build (currently a NOOP).")
    print("\t --branch=<branch> \n\t\t- Build from a specific branch (currently a NOOP).")
    print("\t --rc=<rc number> \n\t\t- Whether or not the build is a release candidate (affects version information).")
    print("\t --iteration=<iteration number> \n\t\t- The iteration to display on the package output (defaults to 0 for RC's, and 1 otherwise).")
    print("\t --race \n\t\t- Whether the produced build should have race detection enabled.")
    print("\t --package \n\t\t- Whether the produced builds should be packaged for the target platform(s).")
    print("\t --nightly \n\t\t- Whether the produced build is a nightly (affects version information).")
    print("\t --parallel \n\t\t- Run Go tests in parallel up to the count specified.")
    print("\t --timeout \n\t\t- Timeout for Go tests. Defaults to 480s.")
    print("\t --clean \n\t\t- Clean the build output directory prior to creating build.")
    print("\t --bucket=<S3 bucket>\n\t\t- Full path of the bucket to upload packages to (must also specify --upload).")
    print("\t --debug \n\t\t- Displays debug output.")
    print("")

def print_package_summary(packages):
    print(packages)

def main():
    # Command-line arguments
    outdir = "build"
    commit = None
    target_platform = None
    target_arch = None
    package_arch = None
    nightly = False
    race = False
    branch = None
    version = get_current_version()
    rc = None
    package = False
    update = False
    clean = False
    upload = False
    test = False
    parallel = None
    timeout = None
    iteration = 1
    no_vet = False
    goarm_version = "6"
    run_get = True
    upload_bucket = None
    global debug

    for arg in sys.argv[1:]:
        if '--outdir' in arg:
            # Output directory. If none is specified, then builds will be placed in the same directory.
            output_dir = arg.split("=")[1]
        if '--commit' in arg:
            # Commit to build from. If none is specified, then it will build from the most recent commit.
            commit = arg.split("=")[1]
        if '--branch' in arg:
            # Branch to build from. If none is specified, then it will build from the current branch.
            branch = arg.split("=")[1]
        elif '--arch' in arg:
            # Target architecture. If none is specified, then it will build for the current arch.
            target_arch = arg.split("=")[1]
        elif '--platform' in arg:
            # Target platform. If none is specified, then it will build for the current platform.
            target_platform = arg.split("=")[1]
        elif '--version' in arg:
            # Version to assign to this build (0.9.5, etc)
            version = arg.split("=")[1]
        elif '--pkgarch' in arg:
            # Package architecture if different from <arch> (armhf, etc)
            package_arch = arg.split("=")[1]
        elif '--rc' in arg:
            # Signifies that this is a release candidate build.
            rc = arg.split("=")[1]
        elif '--race' in arg:
            # Signifies that race detection should be enabled.
            race = True
        elif '--package' in arg:
            # Signifies that packages should be built.
            package = True
        elif '--nightly' in arg:
            # Signifies that this is a nightly build.
            nightly = True
        elif '--upload' in arg:
            # Signifies that the resulting packages should be uploaded to S3
            upload = True
        elif '--parallel' in arg:
            # Set parallel for tests.
            parallel = int(arg.split("=")[1])
        elif '--timeout' in arg:
            # Set timeout for tests.
            timeout = arg.split("=")[1]
        elif '--clean' in arg:
            # Signifies that the outdir should be deleted before building
            clean = True
        elif '--iteration' in arg:
            iteration = arg.split("=")[1]
        elif '--no-vet' in arg:
            no_vet = True
        elif '--goarm' in arg:
            # Signifies GOARM flag to pass to build command when compiling for ARM
            goarm_version = arg.split("=")[1]
        elif '--bucket' in arg:
            # The bucket to upload the packages to, relies on boto
            upload_bucket = arg.split("=")[1]
        elif '--debug' in arg:
            print "[DEBUG] Using debug output"
            debug = True
        elif '--help' in arg:
            print_usage()
            return 0
        else:
            print("!! Unknown argument: {}".format(arg))
            print_usage()
            return 1

    if nightly:
        if rc:
            print("!! Cannot be both nightly and a release candidate! Stopping.")
            return 1
        # In order to support nightly builds on the repository, we are adding the epoch timestamp
        # to the version so that version numbers are always greater than the previous nightly.
        version = "{}.n{}".format(version, int(time.time()))

    # Pre-build checks
    check_environ()
    check_prereqs()

    if not commit:
        commit = get_current_commit(short=True)
    if not branch:
        branch = get_current_branch()
    if not target_arch:
        if 'arm' in get_system_arch():
            # Prevent uname from reporting ARM arch (eg 'armv7l')
            target_arch = "arm"
        else:
            target_arch = get_system_arch()
    if not target_platform:
        target_platform = get_system_platform()
    if rc or nightly:
        # If a release candidate or nightly, set iteration to 0 (instead of 1)
        iteration = 0

    if target_arch == '386':
        target_arch = 'i386'
    elif target_arch == 'x86_64':
        target_arch = 'amd64'

    build_output = {}

    go_get()

    platforms = []
    single_build = True
    if target_platform == 'all':
        platforms = list(supported_builds.keys())
        single_build = False
    else:
        platforms = [target_platform]

    for platform in platforms:
        if platform in prereq_cmds:
            run(prereq_cmds[platform])
        build_output.update( { platform : {} } )
        archs = []
        if target_arch == "all":
            single_build = False
            archs = supported_builds.get(platform)
        else:
            archs = [target_arch]
        for arch in archs:
            od = outdir
            if not single_build:
                od = os.path.join(outdir, platform, arch)
            build(version=version,
                  branch=branch,
                  commit=commit,
                  platform=platform,
                  arch=arch,
                  nightly=nightly,
                  rc=rc,
                  race=race,
                  clean=clean,
                  outdir=od,
                  goarm_version=goarm_version)
            build_output.get(platform).update( { arch : od } )

    # Build packages
    if package:
        if not check_path_for("fpm"):
            print("!! Cannot package without command 'fpm'. Stopping.")
            return 1
        packages = build_packages(build_output, version, package_arch, nightly=nightly, rc=rc, iteration=iteration)
        # Optionally upload to S3
        if upload:
            upload_packages(packages, bucket_name=upload_bucket, nightly=nightly)
    return 0

if __name__ == '__main__':
    sys.exit(main())
