import matplotlib.pyplot as plt
import numpy as np
import csv
import datetime
import sys

dates = []
times = []

def smooth(x, window_len=11):
    s = np.r_[x[window_len-1:0:-1],x,x[-1:-window_len:-1]]
    w = np.hanning(window_len)
    return np.convolve(w/w.sum(),s,mode='valid')

def format_date(x): return datetime.datetime.strptime(x, '%Y-%m-%d %H:%M:%S ')

with open('data/{}.csv'.format(sys.argv[1])) as f:
    reader = csv.reader(f)
    for line in reader:
        try:
            dates.append(format_date(str.split(line[0], '+')[0]))
            times.append(line[1])
        except ValueError:
            pass

dates = np.array(dates, dtype = datetime.date)
times = np.array(times, dtype=int)

print len(times)
print len(dates)

plt.plot(smooth(times, 50))
plt.show()