# %%
import yaml
import os
from datetime import datetime
from pathlib import Path
import reportlab.lib.pagesizes as pagesizes
from reportlab.pdfgen import canvas
from reportlab.platypus import Paragraph
from reportlab.lib.styles import getSampleStyleSheet
import json

""" This script generates a PDF schedule from a YAML file describing a WAP.

TODO: python requirements
pip install reportlab
pip install pyyaml
"""


TEST_PATH = "../data/det6week1.yaml"
SCHEMA_PATH = "../schema/wap.json"
DEFAULT_COLOR = (0.5, 0.5, 0.5, 0.9)


def flip_orientation(orientation: tuple) -> tuple:
    return (orientation[1], orientation[0])


def create_document(data, output_path: os.PathLike = "output.pdf") -> None:
    page_size = flip_orientation(pagesizes.A4)
    output_path = Path(output_path)
    c = canvas.Canvas(str(output_path.absolute()),
                      pagesize=page_size)
    width, height = page_size
    stylesheet = getSampleStyleSheet()
    stylesheet["Code"].fontSize = 7
    title = data["meta"].get("title", "WAP")
    c.drawCentredString(width // 2, height - 20, title)
    if unit := data["meta"].get("unit"):
        c.drawString(40, height - 20, unit)
    if version := data["meta"].get("version"):
        c.drawRightString(width - 20, height - 20, version)

    # metadata
    c.setTitle(title)
    if author := data["meta"].get("author"):
        c.setAuthor(author)

    start_time = parse_time(data["meta"]["startTime"])
    end_time = parse_time(data["meta"]["endTime"])

    # TODO only hour precision for now
    START_HOUR = start_time.hour
    END_HOUR = end_time.hour
    print(f"DEBUG Start hour: {START_HOUR}, End hour: {END_HOUR}")

    category_colors = prepare_categories(data["categories"])

    INNER_ROWS = 2
    INNER_COLS = 6
    ROWS = int(END_HOUR - START_HOUR) * INNER_ROWS
    GRANULARIY = 30  # minutes per row
    COLS = 7 * INNER_COLS
    available_width = int(width - 100)
    available_height = int(height - 100)
    row_height = int(available_height / ROWS)
    col_width = int(available_width / COLS)
    print(f"DEBUG Row height: {row_height}, Column width: {col_width}")
    c.translate(50, 50)
    c.setStrokeColorRGB(0.5, 0.5, 0.5)
    # NOTE there is also canvas.grid
    make_grid(c, ROWS, COLS, row_height, col_width)
    c.setStrokeColorRGB(0.0, 0.0, 1)
    make_grid(c, ROWS // INNER_ROWS, COLS // INNER_COLS,
              row_height * INNER_ROWS, col_width * INNER_COLS)

    # Add time labels
    c.setFont("Helvetica", 8)
    for h, y in zip(range(END_HOUR, START_HOUR-1, -1),
                    range(0, row_height * ROWS + 1, row_height * INNER_ROWS)):
        c.drawString(-20, y - 4, military_time(h, 0))

    def time_to_y(t: datetime.time) -> int:
        return int(t.seconds * row_height / GRANULARIY / 60)

    def draw_event(canvas: canvas.Canvas, event: dict, x: int,
                   column_widths: list) -> None:
        # day -> x
        # start -> y
        # end-start -> height
        # columns -> width
        t_end = parse_time(event["end"])
        t_start = parse_time(event["start"])

        height = time_to_y(t_end - t_start)
        y = time_to_y(end_time - t_end)

        width = 0
        for col, w in column_widths:
            if col in event["columns"]:
                width += w * col_width
                break
            else:
                x += w * col_width
        print(f"DEBUG Drawing event {event['title']} at ({x}, {y}) with size ({width}, {height})")

        category = event.get("category", "default")
        color = category_colors.get(category, DEFAULT_COLOR)
        canvas.setStrokeColorRGB(0, 0, 0, 1)
        canvas.setFillColorRGB(*color)
        canvas.rect(x, y, width, height, fill=1)
        canvas.setFillColorRGB(0, 0, 0, 1)
        text = "<b>" + event["title"] + "</b>"
        if "footnote" not in event:
            if "responsible" in event:
                text += "<br />" + event["responsible"]
            if "location" in event:
                text += ", " + event["location"]
        t = Paragraph(text, stylesheet["Code"])
        # why tf is that needed. Should be exactly the same as the rect above?!
        correction = 2 * col_width
        actual_w, actual_h = t.wrap(width + correction, height)
        t.drawOn(c, x - correction, y + height - actual_h + 1)
        if actual_w > width + correction:
            print(f"WARNING Text too wide: {actual_w} > {width + correction}")
        if actual_h > height:
            print(f"WARNING Text too high: {actual_h} > {height}")

    repeating = []
    for day in data["days"]:
        x = day["offset"] * col_width * INNER_COLS
        cols = columns_per_day(day, repeating)
        column_widths = assign_width(cols, INNER_COLS)
        # Draw the repeating first, the others could overwrite them afterwards
        for event in repeating:
            draw_event(c, event, x, column_widths)
        for event in day["events"]:
            draw_event(c, event, x, column_widths)
            if event.get("repeats", "no") == "daily":
                repeating.append(event)
    c.save()


def prepare_categories(categories: list) -> dict:
    return {cat["identifier"]: parse_color(cat["color"]) for cat in categories}


def parse_color(color: str) -> tuple:
    """Parse a color string into a tuple of normalized RGB values"""
    if color.startswith("#"):
        color = color[1:]
    # RGBA color
    if len(color) == 8:
        return tuple(int(color[i:i+2], 16) / 255 for i in (0, 2, 4, 6))
    # RGB color
    elif len(color) == 6:
        return tuple(int(color[i:i+2], 16) / 255 for i in (0, 2, 4))
    else:
        raise ValueError(f"Invalid color format: {color}")


def columns_per_day(day: dict, repeating: list) -> list:
    """Get all columns appearing in the day"""
    s = set(c for event in day["events"] for c in event["columns"])
    s = s.union(c for event in repeating for c in event["columns"])
    cols = list(sorted(s))
    if "Beso" in cols:
        cols.remove("Beso")
        cols.append("Beso")
    return cols


def assign_width(cols: list, max_cols) -> tuple:
    """Assign widths to each column"""
    # minimum width for each column
    cols = list(cols).copy()
    assert len(cols) <= max_cols, "Too many columns"
    m = list()
    # TODO Beso handling
    max_cols -= 1
    if "Beso" in cols:
        cols.remove("Beso")
    # TODO properly redistribute
    if len(cols) == 0:
        return [("Beso", 1)]
    a, b = divmod(max_cols, len(cols))
    m.extend((c, a + int(b / len(cols)) if i < b else a)
             for i, c in enumerate(cols))
    m.append(("Beso", 1))
    return m


def military_time(hour: int, minute: int) -> str:
    """Convert hour and minute to military time format
    For example, 14:30 -> 1430
    """
    assert 0 <= hour < 24, "Hour must be between 0 and 23"
    assert 0 <= minute < 60, "Minute must be between 0 and 59"
    return f"{hour:02d}{minute:02d}"


def make_grid(c: canvas.Canvas, rows: int, cols: int,
              row_height: int, col_width: int) -> None:
    grid_width = col_width * cols + 1
    grid_height = row_height * rows + 1
    assert grid_width > 0, "Grid width must be positive"
    assert grid_height > 0, "Grid height must be positive"
    # Horizontal lines ---
    for y in range(0, grid_height, row_height):
        c.line(0, y, grid_width, y)
    # Vertical lines |
    for x in range(0, grid_width, col_width):
        c.line(x, 0, x, grid_height)


def parse_time(t: str | int) -> datetime:
    """t is either a string of format hh:mm or an int in minutes"""
    if isinstance(t, (int, float)):
        hours, minutes = divmod(int(t), 60)
        t = f"{hours:02d}:{minutes:02d}"
    return datetime.strptime(t, "%H:%M")


def load_yaml(file_path: os.PathLike) -> dict:
    file_path = Path(file_path)
    if not file_path.exists():
        raise FileNotFoundError(
            f"Yaml WAP file for {file_path} does not exist")
    with open(file_path, "r") as f:
        data = yaml.safe_load(f)
    return data


def load_schema(file_path: os.PathLike) -> tuple[dict, dict]:
    file_path = Path(file_path)
    if not file_path.exists():
        raise FileNotFoundError(
            f"No JSON schema file at {file_path} exists")
    with open(SCHEMA_PATH) as f:
        schema = json.load(f)
    types = extract_types(schema)
    return schema, types


json_to_python_types = {
    "string": str,
    "integer": int,
    "number": float,
    "boolean": bool,
    "array": list,
    "object": dict,
    "null": type(None)
}


# TODO not used yet
# would be nice to get map it directly to python objects and get type hints
def extract_types(schema: dict) -> dict:
    types = {}
    if schema.get("type") == "object":
        for prop, value in schema.get("properties", {}).items():
            if value.get("type") == "array":
                item_type = json_to_python_types.get(
                    value["items"]["type"], object)
                types[prop] = list[item_type]
            else:
                types[prop] = json_to_python_types.get(
                    value.get("type"), object)
    return types


def main():
    schema, _ = load_schema(SCHEMA_PATH)
    data = load_yaml(TEST_PATH)
    create_document(data)


if __name__ == "__main__":
    main()
