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
    public void AddEntry(long t, String n, int axes, float... values){
        try{
            Entries.add(new Entry(Position,t,n,axes,values));
        }
        catch(IllegalArgumentException|SecurityException|IllegalAccessException|NoSuchFieldException ex){
            Log.v("Exception", ex.getMessage());
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
    public long t;
    public String n;
    public float v0;
    public Float v1;
    public Float v2;
    public Float v3;
    public Float v4;
    public Float v5;

    public Entry(String p, long t, String n, int axes, float... values) throws IllegalArgumentException, SecurityException, IllegalAccessException, NoSuchFieldException{
        this.p = p;
        this.t = t;
        this.n = n;

        for (int i = 0; i < axes; i++) {
            Entry.class.getField("v" + i).set(this, values[i]);
        }
    }

}
