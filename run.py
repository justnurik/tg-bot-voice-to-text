import subprocess
import yaml
import os
import sys


def run():
    with open("config.yml") as f:
        config = yaml.safe_load(f)

    args = [
        "./bin/src",
        f"-token={config['api_token']}",
        f"-host-url={config['host_url']}",
        f"-listen-port={config.get('listen_port', 8080)}",
        f"-cache-size={config.get('cache_size', 10000)}",
        f"-log-file={config.get('log_file', 'logs/bot.log')}",
        f"-log-level={config.get('log_level', 'info')}",
        "-model-instance-url=[{}]".format(",".join(config.get("model_instance_urls"))),
    ]

    if config.get("debug", False):
        args.append("-debug")

    if not os.path.exists("./bin/src"):
        print("Compiling Go bot...")
        build = subprocess.run(
            ["go", "build", "-o", "bin/", "-ldflags=-s -w", "./src/..."]
        )
        if build.returncode != 0:
            print("Go build failed.")
            sys.exit(1)

    print("Launching bot...")
    subprocess.run(args)


if __name__ == "__main__":
    run()
