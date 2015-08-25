package de.unibonn.mss15.sensorlogger;

import android.util.Log;
import com.google.gson.Gson;

import java.util.Arrays;
import java.util.List;
import java.util.Vector;


/**
 * Storage structure
 */

public class Storage {
    private transient String Position = "";
    private List<Entry> Entries = new Vector<Entry>();
    private Storage buffer;

    // Append an entry
    public void AddEntry(long t, int e, String n, int axes, float... values){
        try{
            Entries.add(new Entry(Position,t,e,n,axes,values));
        }
        catch(IllegalArgumentException|SecurityException|IllegalAccessException|NoSuchFieldException ex){
            Log.v("Exception", ex.getMessage());
        }
    }

    // Re-set error rates of the last 10 seconds
    public void ResetErrorRates(long stopTime){
        for(int i = Entries.size()-1; i>=0; i--){
            Entry entry = Entries.get(i);
            int diff = (int) (stopTime - entry.t)/1000;
            if( diff < 10 ){
                entry.e = 10 - diff;
                Entries.set(i,entry);
            }
            else
                break;
        }
    }

    // Append a set of entries
    public void Append(Storage s){
        Entries.addAll(s.Entries);
    }

    public void SetPosition(String p){ Position = p; }
    public void Flush(){ Entries.clear(); }
    public void Flush(int from, int to){
        Entries.subList(from,to).clear();
    }
    public int Size(){ return Entries.size(); }

    // Empty constructor
    public Storage(){}

    // Convert storage to json chunks
    public String PopFrontJSON(int chunksize){
        Gson gson = new Gson();
        buffer = new Storage();
        int end = (chunksize>this.Entries.size() ? this.Entries.size() : chunksize);
        // Move some entries to buffer
        buffer.Entries.addAll(this.Entries.subList(0, end)); // copy
        this.Entries.subList(0,end).clear(); // remove

        return gson.toJson(buffer);
    }

    // Save back the buffered entries
    public void Recover(){
        this.Append(buffer);
    }
}


class Entry {
    public String p;
    public long t; // time
    public int e; // error rate
    public String n; // sensor name
    // values
    public float v0;
    public Float v1;
    public Float v2;
    public Float v3;
    public Float v4;
    public Float v5;

    public Entry(String p, long t, int e, String n, int axes, float... values) throws IllegalArgumentException, SecurityException, IllegalAccessException, NoSuchFieldException{
        this.p = p;
        this.t = t;
        this.e = e;
        this.n = n;

        for (int i = 0; i < axes; i++) {
            Entry.class.getField("v" + i).set(this, values[i]);
        }
    }

}
