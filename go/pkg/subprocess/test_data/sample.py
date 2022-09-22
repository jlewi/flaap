import sys
if __name__ == "__main__":
    for i in range(3):
        print(f"stdout Line {i}", file=sys.stdout)
        print(f"stderr Line {i}", file=sys.stderr)