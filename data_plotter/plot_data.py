# -*- coding: utf-8 -*-
"""
Created on Thu Jul  9 18:38:57 2015

@author: omaral-safi
"""

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import urllib.request as urllib
from sys import argv


class Plot:
    def __init__(self, sensor_name):
        #s = pd.Series([1,3,5,np.nan,6,8])
        self.folderName = "plots"
        self.serverURL = "http://46.101.133.187:8529/_db/_system/sensors-data-collector"
        self.preContexts = ["SidePocket", "BackPocket", "Handbag", "SideJacketPocket", "InsideJacketPocket", "InHand", "Idle"]
        self.sensorName = str(sensor_name)
        
        self.getDataAndSave()
        #self.mainData = self.getData()
        #self.plotData()
        #print(dateutil.parser.parse("2015-07-08T20:08:56.405Z"))
        #print(self.mainData.plot(x="time"))
        
        
    def getDataAndSave(self):
        #Get and Plot data for each context
        for context in self.preContexts:
            #get data
            contextData = self.getData(context)
            
            #if data not empty, generate plots
            if not contextData.empty:
               self.savePlot(contextData, context)
            
            
    
    
    def getData(self, context):
        #Get data from the server
        url = self.serverURL + "/list?sensor=" + self.sensorName + "&context=" + context
        req = urllib.Request(url)
        resJSON = urllib.urlopen(req).read()
        
        #Convert data to dataframe
        df = pd.read_json(resJSON)
        
        if not df.empty:
            #Clean data from empty columns
            dataClean = df.dropna(axis='columns', how='all')
            
            #Clean time columns
            dataClean['time'] = dataClean['time'].astype('datetime64[ms]')
            
            #return data
            return dataClean
        
        emptyDf = pd.DataFrame()
            
        return emptyDf
        
        
        
        
        
    def savePlot(self, data, context):
        
        plotTitle = "Sensor: " + self.sensorName + ",  Context: " + context        
        
        #save plot
        ax = data.plot(x="time", title = plotTitle)
        #Construct file name
        fileName = self.sensorName + "_" + context + ".png"
        
        #save to file
        fig = ax.get_figure()
        fig.savefig(self.folderName + "/" + fileName, dpi=200)
        
        
        
        
        
        
        
    
    
    
def main():
    sensor_name = argv[1]
    Plot(sensor_name)
    
if __name__ == '__main__': main()