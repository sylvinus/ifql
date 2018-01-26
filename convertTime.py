import time
from datetime import datetime

f = open("birds.almost","r")

for line in f :
    parts = line.split()
    t = parts[-2] + " " + parts[-1]
    t = t[:-4]
    dt = datetime.strptime(t, "%Y-%m-%d %H:%M:%S")
    print " ".join(parts[:-2]) + " " + str(int(time.mktime(dt.timetuple())))
