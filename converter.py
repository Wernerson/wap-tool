# This script can be used to convert yaml files, that still contain Beso events/columns to the new format, where Beso events are defined in the remarks

from yaml import load, dump, CLoader as Loader, CDumper as Dumper
import argparse

def main():
    parser = argparse.ArgumentParser(description="Process input and output strings.")
    parser.add_argument("-i", '--input', required=True, help='Input file')
    parser.add_argument("-o", '--output', required=True, help='Output file')

    args = parser.parse_args()
    
    data = load(open(args.input, "rb"), Loader=Loader)

    weeks = data["weeks"]
    for week in weeks:
        for day in week["days"]:
            newEvents = []
            if "remarks" not in day.keys():
                day["remarks"] = []
            remarks = day["remarks"]
            if "events" not in day.keys():
                continue
            for event in day["events"]:
                if "appearsIn" in event.keys() and "Beso" in event["appearsIn"]:
                    remarks.append({"title": event["title"] + (" " + event["description"] if "description" in event.keys() else ""), 
                                "start": event["start"], 
                                "end": event["end"]})
                else:
                    newEvents.append(event)
            day["events"] = newEvents
            day["columns"] = [x for x in day["columns"] if x != "Beso"]

    dump(data, open(args.output, "w"), Dumper=Dumper)

if __name__ == "__main__":
    main()