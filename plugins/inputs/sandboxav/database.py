class Measurement:
    def __init__(self):
        pass

    def handleSamples(self, v, colo):
        pass

    def handleTasks(self, v, colo, allsamples):
        pass

    def handleJobs(self, v, colo, allsamples):
        pass

    def updateDatabase(self, influx):
        pass

class SNCounter(Measurement):
    def __init__(self, timestr):
        self.timestr = timestr
        self.dict = {}

    def handleTasks(self, v, colo, allsamples):
        k = v["utm_serial_number"]
        if len(k)>0:
            if k in self.dict:
                self.dict[k] += 1
            else:
                self.dict[k] = 1

    def updateDatabase(self, influx):
        objs = []
        for k in self.dict:
            obj = { "measurement": "tasks_sn",
                    "tags": {"sn": k},
                    "time": self.timestr,
                    "fields": {"value": self.dict[k]}}
            objs.append(obj)
        log.info("inserting tasks_sn")
        influx.write_points(objs)

class MethodCounter(Measurement):
    def __init__(self, timestr):
        self.timestr = timestr
        self.dict = {}

    def handleTasks(self, v, colo, allsamples):
        k = v["method"]
        if len(k)>0:
            if k in self.dict:
                self.dict[k] += 1
            else:
                self.dict[k] = 1

    def updateDatabase(self, influx):
        objs = []
        for k in self.dict:
            obj = { "measurement": "tasks_method",
                    "tags": {"method": k},
                    "time": self.timestr,
                    "fields": {"value": self.dict[k]}}
            objs.append(obj)
        log.info("inserting tasks_method")
        influx.write_points(objs)

class TypeCounter(Measurement):
    def __init__(self, timestr):
        self.timestr = timestr
        self.dict = {}

    def handleSamples(self, v, colo):
        k = v["task_type"]
        if len(k)>0:
            if k in self.dict:
                self.dict[k] += 1
            else:
                self.dict[k] = 1

    def updateDatabase(self, influx):
        objs = []
        for k in self.dict:
            obj = { "measurement": "samples_type",
                    "tags": {"type": k},
                    "time": self.timestr,
                    "fields": {"value": self.dict[k]}}
            objs.append(obj)
        log.info("inserting samples_type")
        influx.write_points(objs)

class ColoCounter(Measurement):
    def __init__(self, timestr):
        self.timestr = timestr
        self.dict = {"sjc":{"submit":0, "unique":0, "submitsize":0, "uniquesize":0, "good":0, "bad":0, "failure":0},
                     "mia":{"submit":0, "unique":0, "submitsize":0, "uniquesize":0, "good":0, "bad":0, "failure":0},
                     "ams":{"submit":0, "unique":0, "submitsize":0, "uniquesize":0, "good":0, "bad":0, "failure":0},
                     "tko":{"submit":0, "unique":0, "submitsize":0, "uniquesize":0, "good":0, "bad":0, "failure":0},
                     "fra":{"submit":0, "unique":0, "submitsize":0, "uniquesize":0, "good":0, "bad":0, "failure":0}}

    def handleSamples(self, v, colo):
        c = self.dict[colo]
        c["unique"] += 1
        c["uniquesize"] += v["file_size"]
        if v["status"] == "good":
            c["good"] += 1
        elif v["status"] == "bad":
            c["bad"] += 1
        elif v["status"] == "failure":
            c["failure"] += 1

    def handleTasks(self, v, colo, allsamples):
        c = self.dict[colo]
        c["submit"] += 1
        if v["sample_sha256"] in allsamples:
            c["submitsize"] += allsamples[v["sample_sha256"]]["file_size"]

    def updateDatabase(self, influx):
        objs = []
        for k in self.dict:
            obj = { "measurement": "colos",
                    "tags": {"colo": k},
                    "time": self.timestr,
                    "fields": self.dict[k]}
            objs.append(obj)
        log.info("inserting colos")
        influx.write_points(objs)

class JobCounter(Measurement):
    def __init__(self, timestr):
        self.timestr = timestr
        self.dict = {}

    def handleJobs(self, v, colo, allsamples):
        k = v["analyze_feature"]
        if len(k) <= 0:
            return

        if k in self.dict:
            o = self.dict[k]
        else:
            o = {"count":1, "sum_pending":0, "sum_running":0, "failure":0, "unknown":0, "good":0, "bad":0}
            self.dict[k] = o

        o["count"] += 1
        o["sum_pending"] += int(v["start_time"]) - int(v["create_time"])
        o["sum_running"] += int(v["finish_time"]) - int(v["start_time"])
        if v["status"] == "failure":
            o["failure"] += 1
        elif v["analyze_result"] == "unknown":
            o["unknown"] += 1
        elif int(v["analyze_result"]) >= 50:
            o["bad"] += 1
        else:
            o["good"] += 1

    def updateDatabase(self, influx):
        objs = []
        for k in self.dict:
            o = self.dict[k]
            obj = { "measurement": "jobs_type",
                    "tags": {"type": k},
                    "time": self.timestr,
                    "fields": {"value":o["count"], "failure":o["failure"], "unknown":o["unknown"], "good":o["good"], "bad":o["bad"],
                               "pending_time":o["sum_pending"]/o["count"], "running_time":o["sum_running"]/o["count"]}}
            objs.append(obj)
        log.info("inserting jobs_type")
        influx.write_points(objs)

counters = [SNCounter(endtimestr), MethodCounter(endtimestr), TypeCounter(endtimestr), ColoCounter(endtimestr), JobCounter(endtimestr)]
allsamples = {}
    for row in samples["data"]:
        v = dict(zip(samples["schema"], row))
        if v['file_size'] != 'None':
            v['file_size'] = int(v['file_size'])
        else:
            v['file_size'] = 0
        allsamples[v["sha256"]] = v
        for counter in counters:
            counter.handleSamples(v, coloname)

    for row in tasks["data"]:
        v = dict(zip(tasks["schema"], row))
        for counter in counters:
            counter.handleTasks(v, coloname, allsamples)

    for row in jobs["data"]:
        v = dict(zip(jobs["schema"], row))
        for counter in counters:
            counter.handleJobs(v, coloname, allsamples)

