#!/usr/bin/env python3

import sys
from os import EX_OK, EX_TEMPFAIL, environ
from subprocess import check_output, call
import os.path
from logging import DEBUG, INFO, StreamHandler, getLogger
from configparser import ConfigParser
from argparse import ArgumentParser

CONFIG_SECTION = 'telegraf-selective-build'
PLUGIN_TYPES = ('inputs', 'outputs', 'aggregators', 'processors')
logger = getLogger(__name__)


def parse_args(args=None):
    parser = ArgumentParser(
        description='build Telegraf with selected group of plugins')
    parser.add_argument(
        '--inputs', help='comma separated list of input plugins to include')
    parser.add_argument(
        '--outputs', help='comma separated list of output plugins to include')
    parser.add_argument(
        '--aggregators',
        help='comma separated list of aggregator plugins to include')
    parser.add_argument(
        '--processors',
        help='comma separated list of processor plugins to include')
    parser.add_argument('--no-pprof',
                        action='store_true',
                        help='disable pprof endpoint')
    parser.add_argument('--config',
                        help='path to .ini config file of the options')
    parser.add_argument('--no-build', action='store_true', help='do not build')
    parser.add_argument('--no-git',
                        action='store_true',
                        help='do not perform git operations')
    parser.add_argument('--verbose',
                        action='store_true',
                        help='print verbose output')
    return parser.parse_args(args)


def get_config(config_path=None):
    """Read the config file and return a dict of configuations"""

    config_parser = ConfigParser()
    if config_path:
        logger.debug('reading config from {} ...'.format(config_path))
        with open(config_path, 'rt') as fh:
            config_parser.read_file(fh)

    return config_parser


def selected_plugins(plugin_type, opts, config):
    """Return a list of plugins of the specifed type, requested via the env/config"""

    plugins = getattr(opts, plugin_type)
    if not plugins or plugins == '':
        env = environ.get('TELEGRAF_{}'.format(plugin_type.upper()),
                          '').strip()
        plugins = config.get(CONFIG_SECTION, plugin_type,
                             fallback='').strip() if env == '' else env
    return [p.strip() for p in plugins.split(',') if p.strip()]


def disable_pprof_endpoint():
    logger.info('disabling pprof endpoint')
    telegraf_go_file = os.path.join('cmd', 'telegraf', 'telegraf.go')
    with open(telegraf_go_file, 'rt') as fh:
        lines = fh.readlines()
    lines = [line for line in lines if r'"net/http/pprof"' not in line]
    with open(telegraf_go_file, 'wt') as fh:
        fh.write(''.join(lines))


def keep_selected_plugins(plugin_type, plugins):
    """For each plugin type, remove all the plugins except the selected ones.
    an empty list keeps all the plugins
    """
    if not plugins or len(plugins) < 1:
        logger.info('keeping all {} plugins'.format(plugin_type))
        return
    logger.info('removing {} plugins except: {}'.format(
        plugin_type, ','.join(plugins)))
    all_plugins_file = os.path.join('plugins', plugin_type, 'all', 'all.go')
    logger.info('removing plugin lines from file {}'.format(all_plugins_file))

    plugin_line_prefix = r'telegraf/plugins/{}'.format(plugin_type)
    selected_plugin_lines = [
        r'{}/{}"'.format(plugin_line_prefix, p) for p in plugins
    ]

    with open(all_plugins_file, 'rt') as fh:
        lines = fh.readlines()

    new_lines = []
    for line in lines:
        if plugin_line_prefix not in line:  # not a plugin reference
            new_lines.append(line)
            continue

        keep_line = False
        for plugin_line in selected_plugin_lines:
            if plugin_line in line:
                keep_line = True
                break
        if keep_line:
            new_lines.append(line)
        else:
            logger.debug('excluding: {}'.format(line.strip()))

    with open(all_plugins_file, 'wt') as fh:
        fh.write(''.join(new_lines))


def build_telegraf():
    logger.info('building telegraf ...')
    call(['make', 'telegraf'])


def git_revert_changes():
    logger.info('detecting local changes ...')
    local_changes = check_output(
        ['git', 'status', '--short',
         '--untracked-files=no']).decode('utf-8').strip()
    local_changes = [l.strip() for l in local_changes.splitlines()]
    checkout_files = []
    for line in local_changes:
        if not line.startswith('M'):
            continue
        file_path = line.split(' ', 1)[-1]
        if os.path.basename(file_path) in ('telegraf.go', 'all.go'):
            logger.debug('detected local changes to {}'.format(file_path))
            checkout_files.append(file_path)

    if checkout_files:
        logger.info('reverting local changes ...')
        git_cmd = ['git', 'checkout', '--']
        git_cmd.extend(checkout_files)
        call(git_cmd)


def main(args=None):
    """Build Telegraf only including specified plugins"""
    opts = parse_args(args)
    logger.setLevel(DEBUG if opts.verbose else INFO)
    logger.addHandler(StreamHandler(sys.stdout))

    config = get_config(opts.config)
    if opts.no_pprof or environ.get('TELEGRAF_NO_PPROF') or config.get(
            CONFIG_SECTION, 'no-pprof', fallback=False):
        disable_pprof_endpoint()

    for plugin_type in PLUGIN_TYPES:
        keep_selected_plugins(plugin_type,
                              selected_plugins(plugin_type, opts, config))

    if not opts.no_build:
        build_telegraf()

    if not opts.no_git:
        git_revert_changes()

    return EX_OK


if __name__ == '__main__':
    try:
        sys.exit(main(sys.argv[1:]))
    except KeyboardInterrupt:
        sys.exit(EX_TEMPFAIL)
