import json, sys, openpyxl, os
import dwn

def exceltojson(inp, out):
    outdict = {}
    wb = openpyxl.load_workbook(inp)
    sheet = wb['Sheet1']
    prefix = "Lök"

    for row in sheet.rows:
        if row[0].value != None and row[1].value == None and row[2].value == None:
            prefix = row[0].value
            continue

        outdict[row[1].value] = prefix + ": " + str(row[2].value)

    f = open(out, "w", encoding="utf8")
    f.write(json.dumps(outdict))
    f.close

def main():
    directory = sys.argv[1]
    filename = directory + "/tmp.xlsx"
    dwn.run(filename)
    exceltojson(filename, directory + "/aboutme.json")
    os.remove(filename)
    print("Done!")


if __name__ == "__main__":
    main()
