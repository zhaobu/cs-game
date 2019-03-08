#!/usr/bin/env bash
fbsPath="/Users/xh/git/maxpanda/protomaxpanda"
targetPath="/Users/xh/gojob/src/mahjong.club/fbs"
subdirList="fbs"


# 删除目标文件下的所有目录
for dir in `ls $targetPath`
do
	if [ -d "$targetPath/$dir" ];then
		#echo "rm -rf $targetPath/$dir"
		rm -rf $targetPath/$dir
	fi
done


echo "cd $fbsPath"
cd $fbsPath

for subdir in $subdirList
do
	# 编译文件
	filelist=`ls $fbsPath/$subdir`

	for file in $filelist
	do
		if [ -d "$fbsPath/$subdir/$file" ];then
			#echo "rm -rf $fbsPath/$subdir/$file"
			rm -rf $fbsPath/$subdir/$file
		elif [ -f $1 ];then
			#echo "flatc --go $subdir/$file"
			flatc --go "$subdir/$file"
		fi
	done
done
echo "编译文件完成..."

# 拷贝到目标文件夹
for subdir in $subdirList
do
filelist=`ls $fbsPath/$subdir`
	#echo $fbsPath/$subdir
	for file in $filelist
	do
		if [ -d "$fbsPath/$subdir/$file" ];then
			#echo "cp -R $fbsPath/$subdir/$file $targetPath/$file"
			cp -R $fbsPath/$subdir/$file $targetPath/$file
			#echo "rm -rf $fbsPath/$subdir/$file"
			rm -rf $fbsPath/$subdir/$file
			
		fi	
	done
done

echo "拷贝文件完成..."

# 安装应用
for subdir in `ls $targetPath`
do
	echo  "go install ../mahjong.club/fbs/$subdir"
	go install ../mahjong.club/fbs/$subdir
done

echo "安装应用完成..."
