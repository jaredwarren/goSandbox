import urllib, re, os, urlparse, json, random, sys, getopt, gzip, math, time
from multiprocessing.dummy import Pool as ThreadPool
from cStringIO import StringIO
from PIL import Image
import numpy as np
 
def download_all_thumbs(ld_num,dest_folder=None):
    
    event_name = 'ludum-dare-%d' % ld_num
    entries_page_url_template = "http://www.ludumdare.com/compo/%s/?action=preview&etype=&start=%%d" % event_name
    
    if dest_folder is None:
        dest_folder = "thumbs/%d/" % ld_num
    if not os.path.exists(dest_folder):
        os.makedirs(dest_folder)
        
    pool = ThreadPool()
        
    def get_games_on_page(url):
        r"""return a list of match objects
            action will be in group 1.
            id will be in group 2.  
            url to thumbnail will be in group 3
            game title will be in group 4
            author will be in group 5"""
        entry_re = re.compile(r"<a href='\?action=(\w+)&uid=([0-9]+)'><img src='([^']*)'><div class='title'><i>(.*?)</i></div>(.*?)</a></div>")
        f = urllib.urlopen(url)
        contents = f.read().decode('utf-8','replace')
        f.close()
        return entry_re.findall(contents)
        
    def download_thumb(entry):
        action,uid,thumburl,title,author = entry
        author = author.encode('ascii','replace')
        title = title.encode('ascii','replace')
        ext = os.path.splitext(urlparse.urlparse(thumburl).path)[1] or ".jpg"
        print "\tfound %s %s's game %s"%(uid,author,title)  
        thumbfile = "%s/%s%s"%(dest_folder,uid,ext)
        if not os.path.exists(thumbfile):
            src = urllib.urlopen(thumburl)
            if src.code != 200:
                print "ERROR downloading %s %s's game %s: %s %s"%(uid,author,title,thumburl,src.code)
                thumbfile = None
            else:
                bytes = src.read()
                src.close()
                with open(thumbfile,"w") as dest:
                    dest.write(bytes)
        return (action,uid,thumburl,title,author,thumbfile)
        
    game_matches = []
    game_count = 0
    while True:
        page = entries_page_url_template % (game_count)
        print "%d: getting games on page: %s"%(game_count,page)
        page_matches = get_games_on_page(page)
        if 0==len(page_matches):
            print "done!",len(game_matches),"games found."
            break
        game_count += len(page_matches)
        game_matches.extend(pool.map(download_thumb,page_matches))
    return game_matches
    
class Game:
    def __init__(self,args):
        action, self.uid, self.thumburl, self.title, self.author, self.thumbfile = args
        if not self.thumbfile:
            print "SKIPPING %s: NO THUMBNAIL"%self
            self.img = None
            self.aspect = 0
        else:
            with open(self.thumbfile) as f:
                self.img = Image.open(StringIO(f.read())) # problem with too-many-files-open errors
            if self.img.mode != "RGB":
                print "CONVERTING %s from %s to RGB"%(self,self.img.mode)
                self.img = self.img.convert("RGB")
            self.aspect = float(self.img.size[1])/(self.img.size[0] or 1)
        self.placed = None
    def compute_mse(self,target_data,target_w,target_h,patch_w,patch_h):
        # TODO multiprocessing? or numpy?
        img = np.int_(self.img.resize((patch_w,patch_h),Image.ANTIALIAS).getdata()).flatten()
        self.mse = [int(((img-tile)**2).sum()) for tile in target_data]
    def __str__(self):
        return "%s %s's game %s"%(self.uid,self.author,self.title)
    
if __name__=="__main__":
    
    opts, args = getopt.getopt(sys.argv[1:],"",
        ("ld-num=","algo=","target-image=","thumb-width=","patch-width=","skip-json="))
    opts = dict(opts)
    ld_num = int(opts.get("--ld-num","30"))
    algo = opts.get("--algo","greedy")
    skip_json = int(opts.get("--skip-json","1"))
    if algo not in ("greedy","timed","test"):
        sys.exit("unsupported algo %s" % algo)
    if len(args) == 1:
        target_filename = args[0]
    elif args:
        sys.exit("unsupported argument %s" % args[0])
    else:
        target_filename = None
    
    index_file = "%d.json"%ld_num
 
    # thumbs not already downloaded?
    if not os.path.exists(index_file):
        index = download_all_thumbs(ld_num)
        with open(index_file,"w") as out:
            json.dump(index,out)
    else:
        # load the index
        with open(index_file) as index:
            index = json.load(index)
     
    # open all the images
    games = filter(lambda x: x.img,map(Game,index))
    print "loaded %d games for ld %d"%(len(games),ld_num)
 
    # load the target image
    target_imagename = opts.get("--target-image","mona_lisa.jpg")
    target = Image.open(target_imagename)
    print "target image %s is %dx%d"%(target_imagename,target.size[0],target.size[1])
    target_prefix = "%d.%s" % (ld_num, os.path.splitext(os.path.basename(target_imagename))[0])
    
    # work out target size
    thumb_aspect = sum(game.aspect for game in games) / len(games)
    patch_w = int(opts.get("--patch-width","10"))
    patch_h = int(float(patch_w)*thumb_aspect)
    print "patches are %dx%d"%(patch_w,patch_h)
    target_w, target_h = target.size
    target_aspect = float(target_w) / target_h
    cols, rows = 1, 1
    while cols*rows < len(games):
        col_asp = float((cols+1)*patch_w) / (math.ceil(float(len(games)) / (cols+1))*patch_h)
        row_asp = float(cols*patch_w) / (math.ceil(float(len(games)) / cols)*patch_h)
        if abs(col_asp-target_aspect) < abs(row_asp-target_aspect):
            cols += 1
        else:
            rows += 1
    target_w = cols * patch_w
    target_h = rows * patch_h
    print "target is %dx%d tiles, %dx%d pixels"%(cols,rows,target_w,target_h)
    print "there are %d tiles and %d images"%(cols*rows,len(games))
    assert cols and rows
    target = target.convert("RGB").resize((target_w,target_h),Image.ANTIALIAS).load()
    target_data = []
    for y in range(rows):
        yofs = y * patch_h
        for x in range(cols):
            xofs = x * patch_w
            target_data.append(np.int_([channel for yy in range(patch_h) for xx in range(patch_w) for channel in target[xofs+xx,yofs+yy]]))
    # compute MSE
    if algo != "test":
        start_time = time.clock()
        mse_file = "%s.mse.json.gz"%target_prefix
        if not os.path.exists(mse_file):
            print "computing Mean Square Error (MSE) for each thumbnail for each tile in the target:"
            for game in games:
                sys.stdout.write(".")
                sys.stdout.flush()
                game.compute_mse(target_data,target_w,target_h,patch_w,patch_h)
            if not skip_json:
                gzip.open(mse_file,"wb",9).write(json.dumps({game.uid:game.mse for game in games}))
        else:
            print "loading MSE matches from file..."
            data = json.loads(gzip.open(mse_file,"rb").read())
            for game in games:
                game.mse = data[game.uid]
        print "took",int(time.clock()-start_time),"seconds"
 
    # work out output size etc
    thumb_w = int(opts.get("--thumb-width","30"))
    thumb_h = int(round(float(thumb_w)*thumb_aspect))
    out_w, out_h = cols*thumb_w, rows*thumb_h
    print "thumbs are %dx%d"%(thumb_w,thumb_h)
    print "output is %dx%d"%(out_w,out_h)
    out = Image.new("RGB",(out_w,out_h))
    
    def done():
        # actually paste them into the output mosaic
        for game in games:
            if not game.placed:
                print "game %s not placed :("%game
            else:
                out.paste(game.img.resize((thumb_w,thumb_h),Image.ANTIALIAS),game.placed)
        # done
        if target_filename:
            target_pre, target_ext = os.path.splitext(os.path.basename(target_filename))
        else:
            target_pre = "%s.%s"%(target_prefix, algo)
            target_ext = ".jpg"
        print "saving %s%s"%(target_pre,target_ext)
        out.save("%s%s"%(target_pre,target_ext))
        if not skip_json:
            print "saving %s.idx.json"%target_pre
            json.dump({game.uid:game.placed for game in games},open("%s.idx.json"%target_pre,"w"))
    
    # place them
    print "%s placement:" % algo
    start_time = time.clock()
    used = {}
    placements = 0
    score = 0
    def place(game,err,xy,symbol="."):
        global placements, score
        x = xy % cols
        y = xy // cols
        game.placed = (x*thumb_w,y*thumb_h)
        used[xy] = (err,game)
        sys.stdout.write(symbol)
        sys.stdout.flush()
        placements += 1
        score += err
    if algo == "test":
        index = range(len(games))
        random.shuffle(index)
        for i,game in enumerate(games):
            place(game,None,index[i])
    elif algo in ("greedy","timed"):
        print "sorting MSE scores..."
        matches = []
        for game in games:
            for i,mse in enumerate(game.mse):
                matches.append((mse,i,game))
        matches = sorted(matches)
        print "took",int(time.clock()-start_time),"seconds"
        start_time = time.clock()
        for err, xy, game in matches:
            if game.placed: continue
            if xy in used: continue
            place(game,err,xy)
    else:
        raise Exception("unsupported algo %s" % algo)
    print " ",placements,"placements made in",int(time.clock()-start_time),"seconds, scoring",score
    
    if algo == "timed":
        while True:
            try:
                start_time = time.clock()
                start_score = score
                placements = 0
                while True:
                    pos_1 = random.randint(1,cols*rows) - 1
                    pos_2 = random.randint(1,cols*rows) - 1
                    if pos_1 not in used or pos_1 == pos_2:
                        continue
                    test = score
                    err_1, game_1 = used[pos_1]
                    test -= err_1
                    test += game_1.mse[pos_2]
                    if pos_2 in used:
                        err_2, game_2 = used[pos_2]
                        test -= err_2
                        test += game_2.mse[pos_1]
                    if test < score:
                        if pos_2 in used:
                            score -= err_2
                            place(game_2,game_2.mse[pos_1],pos_1,"")
                        else:
                            del used[pos_1]
                        score -= err_1
                        place(game_1,game_1.mse[pos_2],pos_2, ".")
                        assert score == test, (score, test)
                    elif pos_2 in used:
                        test -= game_2.mse[pos_1]
                        for i in xrange(10):
                            pos_3 = random.randint(1,cols*rows) - 1
                            if pos_3 in used:
                                test += game_2.mse[pos_3]
                                err_3, game_3 = used[pos_3]
                                test -= game_3.mse[pos_3]
                                test += game_3.mse[pos_1]
                                if test < score:
                                    score -= err_1
                                    place(game_1,game_1.mse[pos_2],pos_2,"")
                                    score -= err_2
                                    place(game_2,game_2.mse[pos_3],pos_3,"")
                                    score -= err_3
                                    place(game_3,game_3.mse[pos_1],pos_1,str(i))
                                    assert score == test, (score, test)
                                    break
                                test -= game_2.mse[pos_3]
                                test += game_3.mse[pos_3]
                                test -= game_3.mse[pos_1]
            except KeyboardInterrupt:
                pass
            print " ",placements,"improvements made in",int(time.clock()-start_time),"seconds, scoring",score,"(%d improvement)"%(start_score-score)
            done()
            while True:
                print "Continue?",
                cmd = raw_input().lower()
                if cmd in ("y", "n"):
                    if cmd == "n":
                        sys.exit()
                    break
                else:
                    print "(y n)?"
    else:
        done()
        