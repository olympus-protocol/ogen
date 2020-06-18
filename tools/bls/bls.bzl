load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

"""
Herumi's BLS library for go depends on
- herumi/mcl
- herumi/bls
- olympus-protocol/bls-go (herumi modified)
"""

def bls_deps():
    _maybe(
        http_archive,
        name = "bls_go",
        strip_prefix = "bls-go-7a5188ecaa2c0764114800baf15b8c5e2154f370",
        urls = [
            "https://github.com/olympus-protocol/bls-go/archive/7a5188ecaa2c0764114800baf15b8c5e2154f370.tar.gz",
        ],
	sha256 = "515fc2ddcd55b83e1f2e83cf16dbc1ff02d3527ff49db4276fdca618600ff6b2",
        build_file = "@ogen//tools/bls:bls_go.BUILD",
    )
    _maybe(
        http_archive,
        name = "mcl",
        strip_prefix = "mcl-66716376b1c48c7aef8f173cac13d1f3b775959d",
        urls = [
            "https://github.com/herumi/mcl/archive/66716376b1c48c7aef8f173cac13d1f3b775959d.tar.gz",
        ],
        sha256 = "e30df1df3b0b9b19dcc814d111e7c485e685329ecf4c846bcfbb91417d12682b",
        build_file = "@ogen//tools/bls:mcl.BUILD",
    )
    _maybe(
        http_archive,
        name = "bls",
        strip_prefix = "bls-1f4204f8b9be007aab3df4b946fa0b952a488afc",
        urls = [
            "https://github.com/herumi/bls/archive/1f4204f8b9be007aab3df4b946fa0b952a488afc.tar.gz",
        ],
        sha256 = "9a6ba9ee95cb7ae8b47366060d7fd186bb48eb9f603c4a54c8ccecc7c2e101f5",
        build_file = "@ogen//tools/bls:bls.BUILD",
    )

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
