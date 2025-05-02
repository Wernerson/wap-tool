import yaml
from pathlib import Path


def minutes_to_hhmm(minutes):
    hours = minutes // 60
    mins = minutes % 60
    return f"{hours:02}:{mins:02}"


# Load the YAML data
input_path = Path("../data/det6.yaml")
output_path = Path("../data/det6_updated.yaml")

try:
    with input_path.open('r') as f:
        data = yaml.safe_load(f)
except Exception as e:
    print(f"Error reading input file: {e}")
    exit(1)

# Modify the structure
for week in data.get("weeks", []):
    for day in week.get("days", []):
        for event in day.get("events", []):
            # Convert Start and End to HH:MM if they are integers
            for key in ["start", "end"]:
                value = event.get(key)
                if isinstance(value, int):
                    event[key] = minutes_to_hhmm(value)
            description_parts = []
            if event.get("responsible") is not None:
                description_parts.append(event["responsible"])
                del event["responsible"]
            if event.get("location") is not None:
                description_parts.append(event["location"])
                del event["location"]
            description = ", ".join(description_parts) if description_parts else None
            if description:
                event["description"] = description

# Write the updated YAML
try:
    with output_path.open('w') as f:
        yaml.safe_dump(data, f, allow_unicode=True, sort_keys=False)
except Exception as e:
    print(f"Error writing output file: {e}")
    exit(1)

print("Migration completed successfully!")
